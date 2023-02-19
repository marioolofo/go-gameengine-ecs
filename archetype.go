package ecs

import (
	"sort"
	"unsafe"
)

type Entity struct {
	EntityID
	arch *Archetype
	row  uint
}

func (e Entity) Get(component EntityID) unsafe.Pointer {
	for i, c := range e.arch.components {
		if c == component {
			return e.arch.columns[i].Get(e.row)
		}
	}
	panic("should never happen")
}

// ArchetypeGraph defines the operations to create and cache the relation between entities and components.
type ArchetypeGraph interface {
	// Add adds the entity in the graph, managing the creation and caching of the relation
	Add(entity EntityID, components ...EntityID)
	// Rem removes the entity and it's components from the graph
	Rem(entity EntityID)
	// AddComponent adds a component to an entity, doing all the transition necessary for the new archetype
	AddComponent(entity, component EntityID)
	// RemComponent removes a component from an entity, doing all the transition necessary for the new archetype
	RemComponent(entity, component EntityID)
	// Get returns the archetype and row for a given entity
	Get(entity EntityID) (*Archetype, uint32)

	Each(components []EntityID, fn func(entity Entity))
}

// ArchEdge defines the link between archetypes.
// the archetypes are connected to form a graph for faster add/removal of components from entities.
type ArchEdge struct {
	add int
	rem int
}

// Archetype contains the definition of a specific entity type.
// it contains the component ids and the storage for the components.
// when a component is a singleton, the Storage is nil and the data is accessed
// by the ComponentFactory.SingletonPtr
type Archetype struct {
	id         uint32
	mask       uint32
	components []EntityID
	columns    []Storage
	edges      map[EntityID]ArchEdge
	entities   []EntityID
}

func (a *Archetype) GetColumnsFor(components ...EntityID) []int {
	result := []int{}

	for _, component := range components {
		index, _ := sort.Find(len(a.components), MakeEntityFindFn(a.components, component))
		result = append(result, index)
	}
	return result
}

func (a *Archetype) GetComponentPtr(col int, row uint32) unsafe.Pointer {
	return a.columns[col].Get(uint(row))
}

// archetypeEntityIndex informs in wich archetype and row the components for the entity is stored.
type archetypeEntityIndex struct {
	archetype int
	row       uint32
}

type archetypeGraph struct {
	factory      ComponentFactory
	entityMap    map[EntityID]archetypeEntityIndex
	archetypeMap map[uint32]int
	archetypes   []Archetype
}

// NewArchetypeGraph returns an ArchetypeGraph responsible for creating and caching the
// relationship between entities and components.
func NewArchetypeGraph(factory ComponentFactory) ArchetypeGraph {
	arch := &archetypeGraph{
		factory,
		make(map[EntityID]archetypeEntityIndex),
		make(map[uint32]int),
		make([]Archetype, 0, 256),
	}
	arch.archetypeMap[0] = arch.newArchetype(0, 0)
	return arch
}

func (a *archetypeGraph) Add(entity EntityID, components ...EntityID) {
	_, exists := a.entityMap[entity]
	if exists {
		panic("trying to add the same entity twice (did you mean AddComponent instead?)")
	}

	archetype := a.findOrCreateArchetype(components)
	row := a.getUnusedRow(archetype, entity)

	a.entityMap[entity] = archetypeEntityIndex{archetype, row}
}

func (a *archetypeGraph) Rem(entity EntityID) {
	cache, ok := a.entityMap[entity]
	if ok {
		a.compressRow(cache.archetype, cache.row)
		delete(a.entityMap, entity)
	}
}

func (a *archetypeGraph) Get(entity EntityID) (*Archetype, uint32) {
	cache, ok := a.entityMap[entity]
	if !ok {
		return nil, 0
	}

	return &a.archetypes[cache.archetype], cache.row
}

func (a *archetypeGraph) AddComponent(entity, component EntityID) {
	cache, ok := a.entityMap[entity]
	if !ok {
		// should panic?
		return
	}

	arch := &a.archetypes[cache.archetype]

	// If already have the component, do nothing
	components := arch.components
	_, found := sort.Find(len(components), MakeEntityFindFn(components, component))
	if found {
		return
	}

	a.updateEntityRelation(entity, component, cache.archetype, cache.row, true)
}

func (a *archetypeGraph) RemComponent(entity, component EntityID) {
	cache, ok := a.entityMap[entity]
	if !ok {
		return
	}

	arch := &a.archetypes[cache.archetype]

	components := arch.components
	_, found := sort.Find(len(components), MakeEntityFindFn(components, component))
	if !found {
		return
	}
	// keep the entity even if it has no components, because it still exists in the graph
	a.updateEntityRelation(entity, component, cache.archetype, cache.row, false)
}

func (a *archetypeGraph) findOrCreateArchetype(components []EntityID) int {
	if len(components) == 0 {
		return 0
	}

	sort.Sort(EntityIDSlice(components))
	key := HashEntityIDArray(components, 0)

	arch, ok := a.archetypeMap[key]
	if !ok {
		arch = a.prepareNewArchetype(key, components)
		a.archetypeMap[key] = arch
	}
	return arch
}

func (a *archetypeGraph) updateEntityRelation(
	entity, component EntityID,
	from int, row uint32, toAdd bool) {

	arch := a.findOrCreateConnection(from, component, toAdd)
	newRow := a.moveEntity(entity, from, arch, row)
	a.entityMap[entity] = archetypeEntityIndex{arch, newRow}
}

func (a *archetypeGraph) findOrCreateConnection(from int, component EntityID, toAdd bool) int {
	fromArch := a.archetypes[from]
	edge, ok := fromArch.edges[component]
	if ok && edge.add > -1 {
		return edge.add
	}

	var components []EntityID
	if toAdd {
		components = make([]EntityID, len(fromArch.components)+1)
		copy(components, fromArch.components)
		components[len(fromArch.components)] = component
		sort.Sort(EntityIDSlice(components))
	} else {
		index, _ := sort.Find(len(fromArch.components), MakeEntityFindFn(fromArch.components, component))
		components = append(fromArch.components[0:index], fromArch.components[index+1:]...)
	}

	key := HashEntityIDArray(components, 0)

	index, ok := a.archetypeMap[key]
	if !ok {
		index = a.prepareNewArchetype(key, components)
		a.archetypeMap[key] = index
	}

	arch := a.archetypes[index]

	if toAdd {
		arch.edges[component] = ArchEdge{add: -1, rem: from}
		fromArch.edges[component] = ArchEdge{add: index, rem: -1}
	} else {
		arch.edges[component] = ArchEdge{add: from, rem: -1}
		fromArch.edges[component] = ArchEdge{add: -1, rem: index}
	}

	return index
}

func (a *archetypeGraph) moveEntity(entity EntityID, from, to int, row uint32) uint32 {
	toColIndex := 0
	toRow := a.getUnusedRow(to, entity)

	fromArch := &a.archetypes[from]
	toArch := &a.archetypes[to]

	for index, col := range fromArch.columns {
		// skip singletons
		if col != nil {
			for toColIndex < len(toArch.columns) && toArch.components[toColIndex] != fromArch.components[index] {
				toColIndex++
			}
			if toColIndex >= len(toArch.columns) {
				break
			}
			toArch.columns[toColIndex].Copy(uint(toRow), col.Get(uint(row)))
		}
	}

	a.compressRow(from, row)

	return toRow
}

func (a *archetypeGraph) newArchetype(key uint32, componentCount int) int {
	index := len(a.archetypes)
	a.archetypes = append(a.archetypes, Archetype{
		id:         key,
		mask:       0,
		components: make([]EntityID, componentCount),
		columns:    make([]Storage, componentCount),
		edges:      make(map[EntityID]ArchEdge),
		entities:   make([]EntityID, 0),
	})

	return index
}

func (a *archetypeGraph) prepareNewArchetype(key uint32, components []EntityID) int {
	index := a.newArchetype(key, len(components))
	arch := &a.archetypes[index]

	for index, comp := range components {
		arch.components[index] = comp
		arch.mask |= (1 << (comp.ID() & 31))
		if !comp.IsSingleton() {
			reg, ok := a.factory.GetByID(comp)
			if !ok {
				panic("trying to use components not registered (did you registered it in the ComponentFactory?)")
			}
			arch.columns[index] = reg.NewStorage()
		}
	}
	return index
}

func (a *archetypeGraph) getUnusedRow(index int, entity EntityID) uint32 {
	arch := &a.archetypes[index]
	row := uint32(len(arch.entities))
	arch.entities = append(arch.entities, entity)
	return row
}

func (a *archetypeGraph) compressRow(index int, row uint32) {
	arch := a.archetypes[index]

	lastRow := uint(len(arch.entities) - 1)
	entity := arch.entities[lastRow]

	for _, col := range arch.columns {
		if col != nil {
			col.Copy(uint(row), col.Get(lastRow))
		}
	}
	arch.entities[row] = entity
	arch.entities = arch.entities[:lastRow]
	cache, _ := a.entityMap[entity]
	cache.row = uint32(row)
	a.entityMap[entity] = cache
}

func (a *archetypeGraph) Each(components []EntityID, fn func(entity Entity)) {
	sort.Sort(EntityIDSlice(components))

	mask := uint32(0)
	for _, c := range components {
		mask |= (1 << (c.ID() & 31))
	}

	total := len(components)
	entity := Entity{}

	for i := 0; i < len(a.archetypes); i++ {
		arch := &a.archetypes[i]
		if len(arch.entities) == 0 {
			continue
		}
		if arch.mask&mask != mask {
			continue
		}
		found := 0
		for _, c := range arch.components {
			if c > components[found] {
				continue
			}
			if components[found] == c {
				found++
			}
		}
		if found != total {
			continue
		}

		entity.arch = arch
		for row, e := range arch.entities {
			entity.EntityID = e
			entity.row = uint(row)
			fn(entity)
		}
	}
}

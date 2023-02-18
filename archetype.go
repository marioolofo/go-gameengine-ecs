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
	add *Archetype
	rem *Archetype
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
	archetype *Archetype
	row       uint32
}

type archetypeGraph struct {
	factory        ComponentFactory
	rootArchetype  *Archetype
	entityMap      map[EntityID]archetypeEntityIndex
	archetypeCache map[uint32]*Archetype
}

// NewArchetypeGraph returns an ArchetypeGraph responsible for creating and caching the
// relationship between entities and components.
func NewArchetypeGraph(factory ComponentFactory) ArchetypeGraph {
	arch := &archetypeGraph{
		factory,
		nil,
		make(map[EntityID]archetypeEntityIndex),
		make(map[uint32]*Archetype),
	}
	arch.rootArchetype = arch.newArchetype(0, 0)
	arch.archetypeCache[0] = arch.rootArchetype
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

	return cache.archetype, cache.row
}

func (a *archetypeGraph) AddComponent(entity, component EntityID) {
	cache, ok := a.entityMap[entity]
	if !ok {
		// should panic?
		return
	}

	// If already have the component, do nothing
	components := cache.archetype.components
	_, found := sort.Find(len(components), MakeEntityFindFn(components, component))
	if found {
		return
	}

	a.updateEntityRelation(entity, component, cache, true)
}

func (a *archetypeGraph) RemComponent(entity, component EntityID) {
	cache, ok := a.entityMap[entity]
	if !ok {
		return
	}

	components := cache.archetype.components
	_, found := sort.Find(len(components), MakeEntityFindFn(components, component))
	if !found {
		return
	}
	// keep the entity even if it has no components, because it still exists in the graph
	a.updateEntityRelation(entity, component, cache, false)
}

func (a *archetypeGraph) findOrCreateArchetype(components []EntityID) *Archetype {
	if len(components) == 0 {
		return a.rootArchetype
	}

	sort.Sort(EntityIDSlice(components))
	key := HashEntityIDArray(components, 0)

	arch, ok := a.archetypeCache[key]
	if !ok {
		arch = a.prepareNewArchetype(key, components)
		a.archetypeCache[key] = arch
	}
	return arch
}

func (a *archetypeGraph) updateEntityRelation(
	entity, component EntityID,
	index archetypeEntityIndex, toAdd bool) {

	arch := a.findOrCreateConnection(index.archetype, component, toAdd)
	index.row = a.moveEntity(entity, index.archetype, arch, index.row)
	index.archetype = arch
	a.entityMap[entity] = index
}

func (a *archetypeGraph) findOrCreateConnection(archetype *Archetype, component EntityID, toAdd bool) *Archetype {
	edge, ok := archetype.edges[component]
	if ok && edge.add != nil {
		return edge.add
	}

	var components []EntityID
	if toAdd {
		components = make([]EntityID, len(archetype.components)+1)
		copy(components, archetype.components)
		components[len(archetype.components)] = component
		sort.Sort(EntityIDSlice(components))
	} else {
		index, _ := sort.Find(len(archetype.components), MakeEntityFindFn(archetype.components, component))
		components = append(archetype.components[0:index], archetype.components[index+1:]...)
	}

	key := HashEntityIDArray(components, 0)

	arch, ok := a.archetypeCache[key]
	if !ok {
		arch = a.prepareNewArchetype(key, components)
		a.archetypeCache[key] = arch
	}

	prev, _ := arch.edges[component]
	if toAdd {
		arch.edges[component] = ArchEdge{add: prev.add, rem: archetype}
		archetype.edges[component] = ArchEdge{add: arch, rem: edge.rem}
	} else {
		arch.edges[component] = ArchEdge{add: archetype, rem: prev.rem}
		archetype.edges[component] = ArchEdge{add: edge.add, rem: arch}
	}

	return arch
}

func (a *archetypeGraph) moveEntity(entity EntityID, from, to *Archetype, row uint32) uint32 {
	toColIndex := 0
	toRow := a.getUnusedRow(to, entity)

	for index, col := range from.columns {
		// skip singletons
		if col != nil {
			for toColIndex < len(to.columns) && to.components[toColIndex] != from.components[index] {
				toColIndex++
			}
			if toColIndex >= len(to.columns) {
				break
			}
			to.columns[toColIndex].Copy(uint(toRow), col.Get(uint(row)))
		}
	}

	a.compressRow(from, row)

	return toRow
}

func (a *archetypeGraph) newArchetype(key uint32, componentCount int) *Archetype {
	return &Archetype{
		id:         key,
		mask:       0,
		components: make([]EntityID, componentCount),
		columns:    make([]Storage, componentCount),
		edges:      make(map[EntityID]ArchEdge),
		entities:   make([]EntityID, 0),
	}
}

func (a *archetypeGraph) prepareNewArchetype(key uint32, components []EntityID) *Archetype {
	arch := a.newArchetype(key, len(components))
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
	return arch
}

func (a *archetypeGraph) getUnusedRow(archetype *Archetype, entity EntityID) uint32 {
	row := uint32(len(archetype.entities))
	archetype.entities = append(archetype.entities, entity)
	return row
}

func (a *archetypeGraph) compressRow(archetype *Archetype, row uint32) {
	lastRow := uint(len(archetype.entities) - 1)
	entity := archetype.entities[lastRow]

	for _, col := range archetype.columns {
		if col != nil {
			col.Copy(uint(row), col.Get(lastRow))
		}
	}
	archetype.entities[row] = entity
	archetype.entities = archetype.entities[:lastRow]
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

	for _, arch := range a.archetypeCache {
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

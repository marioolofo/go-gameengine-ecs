package ecs

import (
	"sort"
	"unsafe"
)

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
	id          uint32
	entityCount uint32
	components  []EntityID
	columns     []Storage
	edges       map[EntityID]ArchEdge
	recycleRows []uint32
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
	row := a.getUnusedRow(archetype)

	a.entityMap[entity] = archetypeEntityIndex{archetype, row}
	archetype.entityCount++
}

func (a *archetypeGraph) Rem(entity EntityID) {
	cache, ok := a.entityMap[entity]
	if ok {
		cache.archetype.recycleRows = append(cache.archetype.recycleRows, cache.row)
		cache.archetype.entityCount--
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
	index.row = a.moveEntity(index.archetype, arch, index.row)
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

func (a *archetypeGraph) moveEntity(from, to *Archetype, row uint32) uint32 {
	toColIndex := 0
	toRow := a.getUnusedRow(to)

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

	to.entityCount++
	from.entityCount--
	from.recycleRows = append(from.recycleRows, row)

	return toRow
}

func (a *archetypeGraph) newArchetype(key uint32, componentCount int) *Archetype {
	return &Archetype{
		id:          key,
		entityCount: 0,
		components:  make([]EntityID, componentCount),
		columns:     make([]Storage, componentCount),
		edges:       make(map[EntityID]ArchEdge),
		recycleRows: make([]uint32, 0),
	}
}

func (a *archetypeGraph) prepareNewArchetype(key uint32, components []EntityID) *Archetype {
	arch := a.newArchetype(key, len(components))
	for index, comp := range components {
		arch.components[index] = comp
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

func (a *archetypeGraph) getUnusedRow(archetype *Archetype) uint32 {
	if len(archetype.recycleRows) > 0 {
		row := archetype.recycleRows[len(archetype.recycleRows)-1]
		archetype.recycleRows = archetype.recycleRows[0 : len(archetype.recycleRows)-1]
		return row
	}
	return archetype.entityCount
}

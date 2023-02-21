package ecs

import (
	"sort"
	"unsafe"
)

type QueryIterator struct {
	archetypes  []Archetype
	columns     []Storage
	mask        Mask
	archIndex   int
	entityIndex int
	entityTotal int
}

func (e *QueryIterator) Prepare(mask Mask, graph *ArchetypeGraph) bool {
	e.archetypes = graph.archetypes
	e.mask = mask
	e.entityIndex = 0
	e.entityTotal = 0
	e.archIndex = -1

	return e.Next()
}

func (e *QueryIterator) Next() bool {
	if e.entityIndex < e.entityTotal {
		e.entityIndex++
		return true
	}

	e.archIndex++
	for e.archIndex < len(e.archetypes) {
		arch := &e.archetypes[e.archIndex]
		if arch.mask.Contains(e.mask) {
			e.entityIndex = -1
			e.entityTotal = len(arch.entities) - 1
			e.columns = arch.columns
			return true
		}
		e.archIndex++
	}
	return false
}

func (e *QueryIterator) Get(component EntityID) unsafe.Pointer {
	return e.columns[component].Get(uint(e.entityIndex))
}

// ArchEdge defines the link between archetypes.
// the archetypes are connected to form a graph for faster add/removal of components from entities.
type ArchEdge struct {
	add int
	rem int
}

type Mask [4]uint64

func MakeMask(bits ...uint64) Mask {
	mask := Mask{}
	for _, bit := range bits {
		mask.Set(bit)
	}
	return mask
}

func (m *Mask) Set(bit uint64) {
	m[bit>>6] |= (1 << (bit & 63))
}

func (m *Mask) Clear(bit uint64) {
	m[bit>>6] &= ^(1 << (bit & 63))
}

func (m Mask) IsSet(bit uint64) bool {
	return m[bit>>6]&(1<<(bit&63)) != 0
}

func (m Mask) Contains(sub Mask) bool {
	return m[0]&sub[0] == sub[0] &&
		m[1]&sub[1] == sub[1] &&
		m[2]&sub[1] == sub[2] &&
		m[3]&sub[1] == sub[3]
}

// Archetype contains the definition of a specific entity type.
// it contains the component ids and the storage for the components.
// when a component is a singleton, the Storage is nil and the data is accessed
// by the ComponentFactory.SingletonPtr
type Archetype struct {
	id         uint32
	mask       Mask
	components []EntityID
	columns    []Storage
	edges      map[EntityID]ArchEdge
	entities   []EntityID
}

func (a *Archetype) GetComponentPtr(col int, row uint32) unsafe.Pointer {
	return a.columns[col].Get(uint(row))
}

// archetypeEntityIndex informs in wich archetype and row the components for the entity is stored.
type archetypeEntityIndex struct {
	archetype int
	row       uint32
}

type ArchetypeGraph struct {
	factory      ComponentFactory
	entityMap    map[EntityID]archetypeEntityIndex
	archetypeMap map[uint32]int
	archetypes   []Archetype
}

// NewArchetypeGraph returns an ArchetypeGraph responsible for creating and caching the
// relationship between entities and components.
func NewArchetypeGraph(factory ComponentFactory) *ArchetypeGraph {
	arch := &ArchetypeGraph{
		factory,
		make(map[EntityID]archetypeEntityIndex),
		make(map[uint32]int),
		make([]Archetype, 0, 256),
	}
	arch.archetypeMap[0] = arch.newArchetype(0, 0)
	return arch
}

func (a *ArchetypeGraph) Add(entity EntityID, components ...EntityID) {
	_, exists := a.entityMap[entity]
	if exists {
		panic("trying to add the same entity twice (did you mean AddComponent instead?)")
	}

	archetype := a.findOrCreateArchetype(components)
	row := a.getUnusedRow(archetype, entity)

	a.entityMap[entity] = archetypeEntityIndex{archetype, row}
}

func (a *ArchetypeGraph) Rem(entity EntityID) {
	cache, ok := a.entityMap[entity]
	if ok {
		a.compressRow(cache.archetype, cache.row)
		delete(a.entityMap, entity)
	}
}

func (a *ArchetypeGraph) Get(entity EntityID) (*Archetype, uint32) {
	cache, ok := a.entityMap[entity]
	if !ok {
		return nil, 0
	}

	return &a.archetypes[cache.archetype], cache.row
}

func (a *ArchetypeGraph) AddComponent(entity, component EntityID) {
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

func (a *ArchetypeGraph) RemComponent(entity, component EntityID) {
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

func (a *ArchetypeGraph) findOrCreateArchetype(components []EntityID) int {
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

func (a *ArchetypeGraph) updateEntityRelation(
	entity, component EntityID,
	from int, row uint32, toAdd bool) {

	arch := a.findOrCreateConnection(from, component, toAdd)
	newRow := a.moveEntity(entity, from, arch, row)
	a.entityMap[entity] = archetypeEntityIndex{arch, newRow}
}

func (a *ArchetypeGraph) findOrCreateConnection(from int, component EntityID, toAdd bool) int {
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

func (a *ArchetypeGraph) moveEntity(entity EntityID, from, to int, row uint32) uint32 {
	toRow := a.getUnusedRow(to, entity)

	fromArch := &a.archetypes[from]
	toArch := &a.archetypes[to]

	for _, comp := range fromArch.components {
		if !toArch.mask.IsSet(comp.ID()) {
			continue
		}
		// skip singletons
		col := fromArch.columns[comp.ID()]
		if col != nil {
			toArch.columns[comp.ID()].Copy(uint(toRow), col.Get(uint(row)))
		}
	}

	a.compressRow(from, row)

	return toRow
}

func (a *ArchetypeGraph) newArchetype(key uint32, componentCount int) int {
	index := len(a.archetypes)
	a.archetypes = append(a.archetypes, Archetype{
		id:         key,
		mask:       Mask{},
		components: make([]EntityID, componentCount),
		columns:    make([]Storage, 256),
		edges:      make(map[EntityID]ArchEdge),
		entities:   make([]EntityID, 0),
	})

	return index
}

func (a *ArchetypeGraph) prepareNewArchetype(key uint32, components []EntityID) int {
	index := a.newArchetype(key, len(components))
	arch := &a.archetypes[index]

	for index, comp := range components {
		arch.components[index] = comp
		arch.mask.Set(comp.ID())
		if !comp.IsSingleton() {
			reg, ok := a.factory.GetByID(comp)
			if !ok {
				panic("trying to use components not registered (did you registered it in the ComponentFactory?)")
			}
			arch.columns[comp.ID()] = reg.NewStorage(100, 0)
		}
	}
	return index
}

func (a *ArchetypeGraph) getUnusedRow(index int, entity EntityID) uint32 {
	arch := &a.archetypes[index]
	row := uint32(len(arch.entities))
	arch.entities = append(arch.entities, entity)
	for _, comp := range arch.components {
		col := arch.columns[comp.ID()]
		if col != nil {
			col.Expand(uint(row + 1))
		}
	}
	return row
}

func (a *ArchetypeGraph) compressRow(index int, row uint32) {
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

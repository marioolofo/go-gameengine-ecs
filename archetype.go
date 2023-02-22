package ecs

import (
	"unsafe"
)

type QueryIterator struct {
	archetypes  []Archetype
	columns     *[MaxComponentCount]Storage
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
	e.archIndex = 0

	return e.Next()
}

func (e *QueryIterator) Next() bool {
	if e.entityIndex < e.entityTotal {
		e.entityIndex++
		return true
	}

	for e.archIndex < len(e.archetypes) {
		arch := &e.archetypes[e.archIndex]
		e.archIndex++
		if arch.mask.Contains(e.mask) {
			e.entityIndex = -1
			e.entityTotal = len(arch.entities) - 1
			e.columns = &arch.columns
			return true
		}
	}
	return false
}

func (e *QueryIterator) Get(component ComponentID) unsafe.Pointer {
	return e.columns[component].Get(uint(e.entityIndex))
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
	mask       Mask
	components []ComponentID
	columns    [MaxComponentCount]Storage
	edges      map[ComponentID]ArchEdge
	entities   []EntityID
}

func (a *Archetype) GetComponentPtr(col ComponentID, row uint32) unsafe.Pointer {
	return a.columns[col].Get(uint(row))
}

// archetypeEntityIndex informs in wich archetype and row the components for the entity is stored.
type archetypeEntityIndex struct {
	archetype int
	row       uint32
}

type ArchetypeGraph struct {
	factory      *ComponentFactory
	entityMap    map[EntityID]archetypeEntityIndex
	archetypeMap map[uint32]int
	archetypes   []Archetype
}

// NewArchetypeGraph returns an ArchetypeGraph responsible for creating and caching the
// relationship between entities and components.
func NewArchetypeGraph(factory *ComponentFactory) *ArchetypeGraph {
	arch := &ArchetypeGraph{
		factory,
		make(map[EntityID]archetypeEntityIndex),
		make(map[uint32]int),
		make([]Archetype, 0, 256),
	}
	arch.archetypeMap[0] = arch.newArchetype(0, Mask{})
	return arch
}

func (a *ArchetypeGraph) Add(entity EntityID, components ...ComponentID) {
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

func (a *ArchetypeGraph) AddComponent(entity EntityID, component ComponentID) {
	cache, ok := a.entityMap[entity]
	if !ok {
		// should panic?
		return
	}

	// If already have the component, do nothing
	if a.archetypes[cache.archetype].mask.IsSet(uint64(component)) {
		return
	}

	a.updateEntityRelation(entity, component, cache.archetype, cache.row, true)
}

func (a *ArchetypeGraph) RemComponent(entity EntityID, component ComponentID) {
	cache, ok := a.entityMap[entity]
	if !ok {
		return
	}

	if !a.archetypes[cache.archetype].mask.IsSet(uint64(component)) {
		return
	}
	// keep the entity even if it has no components, because it still exists in the graph
	a.updateEntityRelation(entity, component, cache.archetype, cache.row, false)
}

func (a *ArchetypeGraph) findOrCreateArchetype(components []ComponentID) int {
	if len(components) == 0 {
		return 0
	}

	mask := MakeComponentMask(components...)
	key := HashUint64Array(0, mask[:])

	arch, ok := a.archetypeMap[key]
	if !ok {
		arch = a.prepareNewArchetype(key, mask)
		a.archetypeMap[key] = arch
	}
	return arch
}

func (a *ArchetypeGraph) updateEntityRelation(
	entity EntityID, component ComponentID,
	from int, row uint32, toAdd bool) {

	arch := a.findOrCreateConnection(from, component, toAdd)
	newRow := a.moveEntity(entity, from, arch, row)
	a.entityMap[entity] = archetypeEntityIndex{arch, newRow}
}

func (a *ArchetypeGraph) findOrCreateConnection(from int, component ComponentID, toAdd bool) int {
	fromArch := a.archetypes[from]
	edge, ok := fromArch.edges[component]
	if ok && edge.add > -1 {
		return edge.add
	}

	mask := fromArch.mask
	if toAdd {
		mask.Set(uint64(component))
	} else {
		mask.Clear(uint64(component))
	}

	key := HashUint64Array(0, mask[:])

	index, ok := a.archetypeMap[key]
	if !ok {
		index = a.prepareNewArchetype(key, mask)
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

	mask := fromArch.mask.And(toArch.mask)

	bit := mask.NextBitSet(0)
	for bit < MaskTotalBits {
		toArch.columns[bit].Copy(uint(toRow), fromArch.columns[bit].Get(uint(row)))
		bit = mask.NextBitSet(bit + 1)
	}

	a.compressRow(from, row)

	return toRow
}

func (a *ArchetypeGraph) newArchetype(key uint32, mask Mask) int {
	index := len(a.archetypes)
	a.archetypes = append(a.archetypes, Archetype{
		id:       key,
		mask:     mask,
		columns:  [MaxComponentCount]Storage{},
		edges:    make(map[ComponentID]ArchEdge, MaxComponentCount),
		entities: make([]EntityID, 0, 1024),
	})

	return index
}

func (a *ArchetypeGraph) prepareNewArchetype(key uint32, mask Mask) int {
	index := a.newArchetype(key, mask)
	arch := &a.archetypes[index]

	bit := mask.NextBitSet(0)
	for bit < MaskTotalBits {
		reg, ok := a.factory.GetByID(ComponentID(bit))
		if !ok {
			panic("trying to use components not registered (did you registered it in the ComponentFactory?)")
		}
		arch.columns[bit] = reg.NewStorage()
		bit = mask.NextBitSet(bit + 1)
	}

	return index
}

func (a *ArchetypeGraph) getUnusedRow(index int, entity EntityID) uint32 {
	arch := &a.archetypes[index]
	row := uint32(len(arch.entities))
	arch.entities = append(arch.entities, entity)

	bit := arch.mask.NextBitSet(0)
	for bit < MaskTotalBits {
		col := arch.columns[bit]
		if col != nil {
			col.Expand(uint(row + 1))
		}
		bit = arch.mask.NextBitSet(bit + 1)
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

package ecs

import (
	"unsafe"
)

/*
ArchetypeGraph defines the interface for an archetype graph.
This interface allows for multiple implementations, for example, for a version
with 256 components and another slower version with unconstrained components.

# Add inserts an EntityID in the graph and reserves memory for it's components

# Rem removes the entity from the graph, recycling the components

# Get returns the archetype and row for the entity entry

AddComponent adds the ComponentID to the entity, moving it to another archetype.

RemComponent removes the ComponentID from the entity, moving it to another archetype.

Query returns a QueryCursor for the mask
*/
type ArchetypeGraph interface {
	Add(EntityID, ...ComponentID)
	Rem(EntityID)
	Get(EntityID) (*Archetype, uint32)
	AddComponent(EntityID, ComponentID)
	RemComponent(EntityID, ComponentID)
	Query(Mask) QueryCursor
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
	mask     Mask
	columns  [MaxComponentCount]Storage
	edges    map[ComponentID]ArchEdge
	entities []EntityID
}

// Component returns the pointer to the component data at col and row in this archetype
func (a *Archetype) Component(col ComponentID, row uint32) unsafe.Pointer {
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
	archetypeMap map[Mask]int
	archetypes   []Archetype
}

// NewarchetypeGraph returns an ArchetypeGraph responsible for creating and caching the
// relationship between entities and components.
func NewArchetypeGraph(factory ComponentFactory) ArchetypeGraph {
	arch := &archetypeGraph{
		factory,
		make(map[EntityID]archetypeEntityIndex),
		make(map[Mask]int),
		make([]Archetype, 0, 256),
	}
	arch.archetypeMap[Mask{}] = arch.newArchetype(Mask{})
	return arch
}

func (a *archetypeGraph) Add(entity EntityID, components ...ComponentID) {
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

func (a *archetypeGraph) AddComponent(entity EntityID, component ComponentID) {
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

func (a *archetypeGraph) RemComponent(entity EntityID, component ComponentID) {
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

func (a *archetypeGraph) Query(mask Mask) QueryCursor {
	var qc QueryCursor
	qc.prepare(mask, a)
	return qc
}

func (a *archetypeGraph) findOrCreateArchetype(components []ComponentID) int {
	if len(components) == 0 {
		return 0
	}

	mask := MakeComponentMask(components...)

	arch, ok := a.archetypeMap[mask]
	if !ok {
		arch = a.prepareNewArchetype(mask)
		a.archetypeMap[mask] = arch
	}
	return arch
}

func (a *archetypeGraph) updateEntityRelation(
	entity EntityID, component ComponentID,
	from int, row uint32, toAdd bool) {

	arch := a.findOrCreateConnection(from, component, toAdd)
	newRow := a.moveEntity(entity, from, arch, row)
	a.entityMap[entity] = archetypeEntityIndex{arch, newRow}
}

func (a *archetypeGraph) findOrCreateConnection(from int, component ComponentID, toAdd bool) int {
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

	index, ok := a.archetypeMap[mask]
	if !ok {
		index = a.prepareNewArchetype(mask)
		a.archetypeMap[mask] = index
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

func (a *archetypeGraph) newArchetype(mask Mask) int {
	index := len(a.archetypes)
	a.archetypes = append(a.archetypes, Archetype{
		mask:     mask,
		columns:  [MaxComponentCount]Storage{},
		edges:    make(map[ComponentID]ArchEdge, MaxComponentCount),
		entities: make([]EntityID, 0, 1024),
	})

	return index
}

func (a *archetypeGraph) prepareNewArchetype(mask Mask) int {
	index := a.newArchetype(mask)
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

func (a *archetypeGraph) getUnusedRow(index int, entity EntityID) uint32 {
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

func (a *archetypeGraph) compressRow(index int, row uint32) {
	arch := &a.archetypes[index]

	lastRow := uint(len(arch.entities) - 1)
	entity := arch.entities[lastRow]

	bit := arch.mask.NextBitSet(0)
	for bit < MaskTotalBits {
		col := arch.columns[bit]
		if col != nil {
			col.Copy(uint(row), col.Get(lastRow))
		}
		bit = arch.mask.NextBitSet(bit + 1)
	}
	arch.entities[row] = entity
	arch.entities = arch.entities[:lastRow]
	cache := a.entityMap[entity]
	cache.row = uint32(row)
	a.entityMap[entity] = cache
}

package ecs

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// ComponentConfig is a struct that defines the component
type ComponentConfig struct {
	ID                         // ID is the identifier of this component
	ForceAlignment int         // ForceAlignment enforces a especific alignment on allocations
	Component      interface{} // Component is the object used to prepare the System for allocations of this type
}

// Filter is the interface used to collect entities that shares the same group of components
//
// The filter automatically is updated when adding or removing entities and components to world
// If you need to add/remove entities or components when iterating the filter,
// make sure to call World.Lock before the update loop, and World.Unlock after the iteration to commit
// the changes to the world and filters
//
// Entities returns a list of entities that contains the group of components informed for World.NewFilter
//
// World returns the world that the entities belongs to
type Filter interface {
	Entities() []Entity
	World() World
}

// World is the interface used to register components, manage entities and allocate components and filters
//
// The World have the limitation of total MaskTotalBits different components
//
// RegisterComponents create Systems to control the allocation of components defined by the ComponentConfig
// If multiple configs references the same ID, the last will replace the previous System
//
// NewEntity returns a new entity with zero components
//
// RemEntity removes the entity from the World and recycles the components
//
// Assign allocate and maps components to entity
// If the entity already have some component or the ID is invalid, the allocation will be skipped
//
// Remove remove and recycles the components from entity
//
// Component returns the pointer of component for entity if it exists, nil otherwise
//
// Lock defer world updates to filters until Unlock is called
// References for components added to world inside locked scope are valid to be used and stored
//
// Unlock collect the world changes and update the filters
// World only unlocks when you call this function the same amount of times you called Lock
//
// NewFilter returns a filter that contains all entities with the desired components
// The filter is automatically updated when new entities or components are added/removed from the World
//
// RemFilter removes the filter from the World
// The filter may be used after the removal, but the data may become outdated and invalid after world updates
//
// EntityStats collects entity statistics from the World
// ComponentStats collects component statistics from the World
// Stats collects all statistics from the World
type World interface {
	RegisterComponents(configs ...ComponentConfig)

	NewEntity() Entity
	RemEntity(e Entity)

	Assign(entity Entity, components ...ID)
	Remove(entity Entity, components ...ID)
	Component(e Entity, component ID) unsafe.Pointer

	Lock()
	Unlock()

	NewFilter(components ...ID) Filter
	RemFilter(filter Filter)

	EntityStats() WorldEntityStats
	ComponentStats(component ID) SystemStats
	Stats() WorldStats
}

// WorldEntityStats is the runtime statistics of entities in the World
type WorldEntityStats struct {
	Total    uint   // Total entities allocated
	InUse    uint   // Total entities currently in use
	Recycled uint   // Total entities in recycle list
}

// WorldStatus gives real time statistics of the World
type WorldStats struct {
	EntityStats    WorldEntityStats
	ComponentStats [MaskTotalBits]SystemStats
	ComponentCount uint // Number of valid components in ComponentStats array
	Lock           int  // > 0 if the World is inside locked state
}

func (e WorldEntityStats) String() string {
	return fmt.Sprintf("stats.entity\n\t%d Total\n\t%d Used\n\t%d Recycled\n", e.Total, e.InUse, e.Recycled)
}

func (s WorldStats) String() string {
	var builder strings.Builder

	builder.WriteString("World.Stats\n")
	builder.WriteString(fmt.Sprintf("\tworld locked count: %d\n", s.Lock))
	builder.WriteString(s.EntityStats.String())

	for i := uint(0); i < s.ComponentCount; i++ {
		comp := s.ComponentStats[i]

		builder.WriteString(fmt.Sprintf("stats.component %d", comp.MemStats.ID))
		builder.WriteString(fmt.Sprintf("\n\tSparseArray length: %d", comp.SparseArrayLength))
		builder.WriteString(fmt.Sprintf("\n\tMemPool components in use: %d", comp.MemStats.InUse))
		builder.WriteString(fmt.Sprintf("\n\tMemPool components recycled: %d", comp.MemStats.Recycled))
		builder.WriteString(fmt.Sprintf("\n\tMemPool component buffer size: %d", comp.MemStats.SizeInBytes))
		builder.WriteString(fmt.Sprintf("\n\tMemPool component buffer alignment: %d", comp.MemStats.Alignment))
		builder.WriteString(fmt.Sprintf("\n\tMemPool page size: %d", comp.MemStats.PageSizeInBytes))
		builder.WriteString(fmt.Sprintf("\n\tMemPool page count: %d\n", comp.MemStats.TotalPages))
	}
	return builder.String()
}

type EntityMaskPair struct {
	entity ID
	mask   Mask
}

type world struct {
	systemsMask        Mask
	recycleIDs         []ID
	entities           []Mask
	systems            [MaskTotalBits]System
	filters            []*entityFilter
	entityRemoveIndex  []int
	lock               int
	remEntitiesQueue   []Entity
	eventsQueue        []EntityMaskPair
	eventsQueueAdding  []bool
}

// NewWorld returns a new World with Systems created for configs
// New configs can be added at latter stages with World.RegisterComponents
func NewWorld(configs ...ComponentConfig) World {
	w := &world{
		0,
		make([]ID, 0, InitialEntityRecycleCapacity),
		make([]Mask, 1, InitialEntityCapacity),
		[MaskTotalBits]System{},
		make([]*entityFilter, 0),
		make([]int, 0, 100),
		0,
		make([]Entity, 0),
		make([]EntityMaskPair, 0, 100),
		make([]bool, 0, 100),
	}

	w.RegisterComponents(configs...)
	return w
}

func (w *world) RegisterComponents(configs ...ComponentConfig) {
	for _, config := range configs {
		if config.ID < 0 || config.ID >= MaskTotalBits {
			LogMessage("[World.RegisterComponents] bit mask out of range (trying to use bit %d)\n", config.ID)
			continue
		}

		if w.systems[config.ID] != nil {
			LogMessage("[World.AddSystems] trying to use bit %d for %s, but bit is already in use\n", config.ID, reflect.TypeOf(config.Component).Name())
		}
		w.systems[config.ID] = NewSystem(config.ID, config.Component, config.ForceAlignment)
		w.systemsMask.Set(config.ID, true)
	}
}

func (w *world) NewEntity() Entity {
	var id ID

	// It's safe to use the recycleIDs entities even if the world is locked, as it contains
	// IDs removed before the Lock was called
	if len(w.recycleIDs) > 0 {
		id = w.recycleIDs[len(w.recycleIDs)-1]
		w.recycleIDs = w.recycleIDs[:len(w.recycleIDs)-1]
	} else {
		id = ID(len(w.entities))
		w.entities = append(w.entities, 0)
	}

	return id
}

func (w *world) RemEntity(entity Entity) {
	if entity < 1 {
		LogMessage("[World.RemEntity] invalid entity id %d\n", entity)
		return
	}
	if !w.assertEntityExists(
		entity,
		"Trying to re entity that's not in use (entity %d)\n",
		"Trying to remove entity that's already removed inside locked scope (entity %d)\n") {
		return
	}

	// We recycle the entities only when the world is unlocked to prevent
	// invalidate the references to components when iterating filters
	if w.lock > 0 {
		w.remEntitiesQueue = append(w.remEntitiesQueue, entity)
		w.eventsQueue = append(w.eventsQueue, EntityMaskPair{entity, w.entities[entity]})
		w.eventsQueueAdding = append(w.eventsQueueAdding, false)
		w.entities[entity] = Mask(0);
	} else {
		w.remComponentsFromEntities(entity)
		w.recycleEntitiesAndUpdateFilters(entity)
	}
}

func (w *world) Assign(entity Entity, ids ...ID) {
	if entity < 1 {
		LogMessage("[World.Assign] invalid entity id %d\n", entity)
		return
	}
	if !w.assertEntityExists(
		entity,
		"Trying to assign components to invalid entity (entity %d)\n",
		"Trying to assign components to invalid entity in locked scope (entity %d)\n") {
		return
	}

	w.assign(entity, ids...)

	if w.lock > 0 {
		w.eventsQueue = append(w.eventsQueue, EntityMaskPair{entity, w.entities[entity]})
		w.eventsQueueAdding = append(w.eventsQueueAdding, true)
	} else {
		w.updateFilters(true, EntityMaskPair{entity, w.entities[entity]})
	}
}

func (w *world) Remove(entity Entity, ids ...ID) {
	if entity < 1 {
		LogMessage("[World.Remove] invalid entity id %d\n", entity)
		return
	}
	if !w.assertEntityExists(
		entity,
		"Trying to remove components from invalid entity (entity %d)\n",
		"Trying to remove components from invalid entity in locked scope (entity %d)\n") {
		return
	}

	if w.lock > 0 {
		mask := NewMask(ids...)
		w.eventsQueue = append(w.eventsQueue, EntityMaskPair{entity, mask})
		w.eventsQueueAdding = append(w.eventsQueueAdding, false)
		w.entities[entity] &= ^mask
	} else {
		w.remove(entity, NewMask(ids...))
	 	w.updateFilters(false, EntityMaskPair{entity, w.entities[entity]})
	}
}

func (w *world) Component(entity Entity, compID ID) unsafe.Pointer {
	if entity < 1 {
		LogMessage("[World.Component] invalid entity id %d\n", entity)
		return nil
	}
	if compID < 0 || compID >= MaskTotalBits || w.systems[compID] == nil {
		return nil
	}
	return w.systems[compID].Get(entity)
}

func (w *world) Lock() {
	w.lock++
}

func (w *world) Unlock() {
	if w.lock > 0 {
		w.lock--
		if w.lock == 0 {
			w.updateFiltersAfterUnlock()
		}
	} else {
		LogMessage("[World.Unlock] world already unlocked!")
	}
}


func (w *world) EntityStats() WorldEntityStats {
	total := uint(len(w.entities)) - 1 // -1 because index 0 is unused
	recycled := uint(len(w.recycleIDs))

	return WorldEntityStats{
		Total: total,
		InUse: total - recycled,
		Recycled: recycled,
	}
}

func (w *world) ComponentStats(id ID) SystemStats {
	if id < 0 || id >= MaskTotalBits {
		LogMessage("[World.ComponentStats] invalid component id %d\n", id)
		return SystemStats{}
	}

	return w.systems[id].Stats()
}

func (w *world) Stats() WorldStats {
	stats := WorldStats{
		Lock: w.lock,
		EntityStats: w.EntityStats(),
	}
	stats.ComponentCount = 0

	lastBit := w.systemsMask.NextBitSet(0)
	for lastBit < MaskTotalBits {
		stats.ComponentStats[stats.ComponentCount] = w.ComponentStats(ID(lastBit))
		stats.ComponentCount++
		lastBit = w.systemsMask.NextBitSet(lastBit + 1)
	}

	return stats
}

func (w *world) remComponentsFromEntities(entities ...Entity) {
	for _, entity := range entities {
		mask := w.entities[entity]
		lastBit := mask.NextBitSet(0)
		for lastBit < MaskTotalBits {
			w.systems[lastBit].Recycle(entity)
			lastBit = mask.NextBitSet(lastBit + 1)
		}
	}
}

func (w *world) assign(entity Entity, ids ...ID) {
	mask := w.entities[entity]
	for _, id := range ids {
		if id < 0 || id >= MaskTotalBits {
			LogMessage("[World.Assign] invalid component id %d for entity %d\n", id, entity)
			return
		}
		if w.systems[id] == nil {
			LogMessage("[World.Assign] component id %d not registered in this world instance\n", id)
			return
		}
		w.systems[id].New(entity)
		mask.Set(id, true)
	}
	w.entities[entity] = mask
}

func (w *world) remove(entity Entity, removeMask Mask) {
	lastBit := removeMask.NextBitSet(0)
	for lastBit < MaskTotalBits {
		if w.systems[lastBit] != nil {
			w.systems[lastBit].Recycle(entity)
		}
		lastBit = removeMask.NextBitSet(lastBit + 1)
	}
	w.entities[entity] &= ^removeMask
}

func (w *world) recycleEntitiesAndUpdateFilters(entities ...Entity) {
	for _, entity := range entities {
		w.updateFilters(false, EntityMaskPair{entity, w.entities[entity]})
		w.entities[entity] = Mask(0)
	}
	w.recycleIDs = append(w.recycleIDs, entities...)
}

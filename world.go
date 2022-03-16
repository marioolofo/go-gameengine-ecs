package ecs

import (
	"reflect"
	"sort"
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
// NewFilter returns a filter that contains all entities with the desired components
// The filter is automatically updated when new entities or components are added/removed from the World
//
// RemFilter removes the filter from the World
// The filter may be used after the removal, but the data may become outdated and invalid after world updates
type World interface {
	RegisterComponents(configs ...ComponentConfig)

	NewEntity() Entity
	RemEntity(e Entity)

	Assign(entity Entity, components ...ID)
	Remove(entity Entity, components ...ID)
	Component(e Entity, component ID) unsafe.Pointer

	NewFilter(components ...ID) Filter
	RemFilter(filter Filter)
}

type Entities []Entity

type entityFilter struct {
	mask     Mask
	world    World
	entities Entities
}

func (e Entities) Len() int           { return len(e) }
func (e Entities) Less(a, b int) bool { return e[a] < e[b] }
func (e Entities) Swap(a, b int)      { e[a], e[b] = e[b], e[a] }

func (e *entityFilter) Entities() []Entity {
	return e.entities
}

func (e *entityFilter) World() World {
	return e.world
}

type world struct {
	systemsMask Mask
	recycleIDs  []ID
	entities    []Mask
	systems     [MaskTotalBits]System
	filters     []*entityFilter
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
	mask := w.entities[entity]
	lastBit := mask.NextBitSet(0)
	for lastBit < MaskTotalBits {
		w.systems[lastBit].Recycle(entity)
		lastBit = mask.NextBitSet(lastBit + 1)
	}
	w.removeAndRecycleEntity(entity)
}

func (w *world) Assign(entity Entity, ids ...ID) {
	if entity < 1 {
		LogMessage("[World.Assign] invalid entity id %d\n", entity)
		return
	}
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
	w.updateFilters(entity, mask, true)
}

func (w *world) Remove(entity Entity, ids ...ID) {
	if entity < 1 {
		LogMessage("[World.Remove] invalid entity id %d\n", entity)
		return
	}
	mask := w.entities[entity]
	for _, id := range ids {
		if w.systems[id] != nil {
			w.systems[id].Recycle(id)
		}
		mask.Set(id, false)
	}
	w.entities[entity] = mask
	w.updateFilters(entity, mask, false)
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

func (w *world) removeAndRecycleEntity(id ID) {
	w.recycleIDs = append(w.recycleIDs, id)
	w.updateFilters(id, w.entities[id], false)
	w.entities[id] = Mask(0)
}

func (w *world) NewFilter(ids ...ID) Filter {
	filter := &entityFilter{
		mask:     NewMask(ids...),
		entities: make([]Entity, 0),
		world:    w,
	}

	w.collectEntities(filter)

	w.filters = append(w.filters, filter)
	return filter
}

func (w *world) RemFilter(filter Filter) {
	for index, f := range w.filters {
		if f == filter {
			w.filters[index] = w.filters[len(w.filters)-1]
			w.filters = w.filters[:len(w.filters)-1]
			return
		}
	}
}

func (w *world) collectEntities(filter *entityFilter) {
	for index, mask := range w.entities[1:] {
		if mask.Contains(filter.mask) {
			filter.entities = append(filter.entities, Entity(index+1))
		}
	}
	sort.Sort(filter.entities)
}

func (w *world) updateFilters(entity Entity, mask Mask, add bool) {
	for _, filter := range w.filters {
		if mask.Contains(filter.mask) {
			index := sort.Search(len(filter.entities), func(ind int) bool { return filter.entities[ind] == entity })
			if add {
				if index == len(filter.entities) {
					filter.entities = append(filter.entities, entity)
					sort.Sort(filter.entities)
				}
			} else {
				if index != filter.entities.Len() {
					filter.entities[index] = filter.entities[filter.entities.Len()-1]
					filter.entities = filter.entities[:filter.entities.Len()-1]
					sort.Sort(filter.entities)
				}
			}
		}
	}
}

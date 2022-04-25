package ecs

import "sort"

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

func (e *entityFilter) sort() {
	sort.Sort(e.entities)
}

func (e *entityFilter) indexOf(entity Entity, limit int) int {
	index := sort.Search(limit, func(ind int) bool { return e.entities[ind] >= entity })
	if index < limit && e.entities[index] == entity {
		return index
	}
	return limit
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

func (w *world) updateFiltersAfterUnlock() {
	// Only add entities to list of recycled IDs, as the components to remove are in
	// the eventsQueue:
	if len(w.remEntitiesQueue) > 0 {
		w.recycleIDs = append(w.recycleIDs, w.remEntitiesQueue...)
		w.remEntitiesQueue = w.remEntitiesQueue[:0]
	}

	if len(w.eventsQueue) == 0 {
		return
	}
	// components added/removed need to be executed in order to prevent
	// remove/add in wrong order
	endIndex := 0

	for endIndex < len(w.eventsQueue) {
		startIndex := endIndex
		adding := w.eventsQueueAdding[startIndex]

		for adding == w.eventsQueueAdding[endIndex] {
			endIndex++
			if endIndex == len(w.eventsQueue) {
				break
			}
		}
		w.updateFilters(adding, w.eventsQueue[startIndex:endIndex]...)
	}
	w.eventsQueue = w.eventsQueue[:0]
	w.eventsQueueAdding = w.eventsQueueAdding[:0]
}

func (w *world) collectEntities(filter *entityFilter) {
	for index, mask := range w.entities[1:] {
		if mask.Contains(filter.mask) {
			filter.entities = append(filter.entities, Entity(index+1))
		}
	}
	filter.sort()
}

func (w *world) updateFilters(add bool, entityPairs ...EntityMaskPair) {
	if add {
		for _, filter := range w.filters {
			if w.batchInsertToFilter(filter, entityPairs...) {
				filter.sort()
			}
		}
	} else {
		for _, filter := range w.filters {
			if w.batchRemoveFromFilter(filter, entityPairs...) {
				filter.sort()
			}
		}
	}
}

func (w *world) batchRemoveFromFilter(filter *entityFilter, entityPairs ...EntityMaskPair) (needSort bool) {
	needSort = false
	filterEntityCount := len(filter.entities)

	w.entityRemoveIndex = w.entityRemoveIndex[:0]

	for _, entityPair := range entityPairs {
		mask := entityPair.mask
		entity := entityPair.entity

		if mask.Contains(filter.mask) {
			index := filter.indexOf(entity, filterEntityCount)
			if index != filterEntityCount {
				w.entityRemoveIndex = append(w.entityRemoveIndex, index)
			}
		}
	}

	if len(w.entityRemoveIndex) > 0 {
		for sub, index := range w.entityRemoveIndex {
			filter.entities[index] = filter.entities[len(filter.entities) - 1 - sub]
		}
		filter.entities = filter.entities[:len(filter.entities) - len(w.entityRemoveIndex)]
		needSort = true
	}

	return
}

func (w *world) batchInsertToFilter(filter *entityFilter, entityPairs ...EntityMaskPair) bool {
	needSort := false
	filterEntityCount := len(filter.entities)

	for _, entityPair := range entityPairs {
		entity := entityPair.entity
		mask := entityPair.mask

		if mask.Contains(filter.mask) {
			index := filter.indexOf(entity, filterEntityCount)
			if index == filterEntityCount {
				filter.entities = append(filter.entities, entity)
				needSort = true
			}
		}
	}
	return needSort
}

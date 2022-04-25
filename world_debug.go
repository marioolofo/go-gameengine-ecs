package ecs

func (w *world) assertEntityExists(entity Entity, removedEntityText, lockedScopeText string) bool {
	if WorldAssertActions {
		if w.entities[entity] == Mask(0) && entitySliceContains(w.recycleIDs, entity) {
			LogMessage(removedEntityText, entity)
			return false
		}
		if w.lock > 0 && entitySliceContains(w.remEntitiesQueue, entity) {
			LogMessage(lockedScopeText, entity)
			return false
		}
	}
	return true
}

func entitySliceContains(entities []Entity, entity Entity) bool {
	for _, e := range entities {
		if e == entity {
			return true
		}
	}
	return false
}

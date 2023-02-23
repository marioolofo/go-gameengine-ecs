package ecs

import "unsafe"

/*
QueryCursor holds the data to iterate over the entities found for a given mask

Use this implementation in critical parts of your project
to access the entity ID and components in an efficient way
*/
type QueryCursor struct {
	archetypes  []Archetype
	arch        *Archetype
	mask        Mask
	archIndex   int
	entityIndex int
	entityTotal int
}

// Next returns true if the query have more entities to iterate over
func (e *QueryCursor) Next() bool {
	if e.entityIndex < e.entityTotal {
		e.entityIndex++
		return true
	}

	for e.archIndex < len(e.archetypes) {
		arch := &e.archetypes[e.archIndex]
		e.archIndex++
		if len(arch.entities) > 0 && arch.mask.Contains(e.mask) {
			e.entityIndex = 0
			e.entityTotal = len(arch.entities) - 1
			e.arch = arch
			return true
		}
	}
	return false
}

// Component returns the component pointer for the actual entity
func (e *QueryCursor) Component(component ComponentID) unsafe.Pointer {
	return e.arch.columns[component].Get(uint(e.entityIndex))
}

// Entity returns the EntityID of the actual entity
func (e *QueryCursor) Entity() EntityID {
	return e.arch.entities[e.entityIndex]
}

// Restart initializes the cursor to the first entity in the query
func (e *QueryCursor) Restart() {
	e.entityIndex = 0
	e.entityTotal = 0
	e.archIndex = 0
}

func (e *QueryCursor) prepare(mask Mask, graph *archetypeGraph) {
	e.archetypes = graph.archetypes
	e.mask = mask
	e.Restart()
}

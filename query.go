package ecs

import "unsafe"

type QueryCursor struct {
	archetypes  []Archetype
	arch        *Archetype
	mask        Mask
	archIndex   int
	entityIndex int
	entityTotal int
}

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

func (e *QueryCursor) Component(component ComponentID) unsafe.Pointer {
	return e.arch.columns[component].Get(uint(e.entityIndex))
}

func (e *QueryCursor) Entity() EntityID {
	return e.arch.entities[e.entityIndex]
}

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

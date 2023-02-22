package ecs

import "unsafe"

type World interface {
	NewEntity(components ...ComponentID) EntityID
	GetComponent(entity EntityID, component ComponentID) unsafe.Pointer
	Register(comp ComponentRegistry)
	Query(mask Mask) QueryIterator
	GetEntityPool() EntityPool
}

type world struct {
	entityPool EntityPool
	factory    ComponentFactory
	archGraph  *ArchetypeGraph
}

func NewWorld(entityPoolSize uint) World {
	w := &world{
		NewEntityPool(entityPoolSize),
		NewComponentFactory(),
		nil,
	}
	w.archGraph = NewArchetypeGraph(&w.factory)
	return w
}

func (w *world) NewEntity(comp ...ComponentID) EntityID {
	id := w.entityPool.New()
	w.archGraph.Add(id, comp...)
	return id
}

func (w *world) GetComponent(entity EntityID, component ComponentID) unsafe.Pointer {
	arch, row := w.archGraph.Get(entity)
	return arch.columns[component].Get(uint(row))
}

func (w *world) Register(comp ComponentRegistry) {
	w.factory.Register(comp)
}

func (w *world) Query(mask Mask) QueryIterator {
	var qi QueryIterator
	qi.Prepare(mask, w.archGraph)
	return qi
}

func (w *world) GetEntityPool() EntityPool {
	return w.entityPool
}

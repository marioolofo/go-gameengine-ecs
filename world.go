package ecs

import "unsafe"

type World interface {
	NewEntity(...ComponentID) EntityID
	RemEntity(EntityID)
	IsAlive(EntityID) bool

	AddComponent(EntityID, ComponentID)
	RemComponent(EntityID, ComponentID)
	Component(EntityID, ComponentID) unsafe.Pointer

	Register(ComponentRegistry)
	Query(Mask) QueryCursor
}

type WorldConfig struct {
	EntityPool
	ComponentFactory
	ArchetypeGraph
}

type world struct {
	entityPool EntityPool
	factory    ComponentFactory
	archGraph  ArchetypeGraph
}

func NewWorld(entityPoolSize uint) World {
	factory := NewComponentFactory()
	w := &world{
		NewEntityPool(entityPoolSize),
		factory,
		NewArchetypeGraph(factory),
	}
	return w
}

func (w *world) NewEntity(comp ...ComponentID) EntityID {
	id := w.entityPool.New()
	w.archGraph.Add(id, comp...)
	return id
}

func (w *world) RemEntity(id EntityID) {
	w.archGraph.Rem(id)
	w.entityPool.Recycle(id)
}

func (w *world) IsAlive(id EntityID) bool {
	return w.entityPool.IsAlive(id)
}

func (w *world) AddComponent(id EntityID, component ComponentID) {
	w.archGraph.AddComponent(id, component)
}

func (w *world) RemComponent(id EntityID, component ComponentID) {
	w.archGraph.RemComponent(id, component)
}

func (w *world) Component(entity EntityID, component ComponentID) unsafe.Pointer {
	arch, row := w.archGraph.Get(entity)
	return arch.columns[component].Get(uint(row))
}

func (w *world) Register(comp ComponentRegistry) {
	w.factory.Register(comp)
}

func (w *world) Query(mask Mask) QueryCursor {
	return w.archGraph.Query(mask)
}

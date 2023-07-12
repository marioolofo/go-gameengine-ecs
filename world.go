/*
Package ECS implements the Entity-Component-System pattern

Besides the System in the name, this package offers a Query function
to iterate over the needed entities, leaving the systems implementation
to the user.

This implementation is modular, so you can create the ComponentRegistry directly,
instantiate the EntityPool to control the entities alive in the project and use
the ArchetypeGraph to fast query the graph for specific entities.

For more, see the examples in the example folder
*/
package ecs

import "unsafe"

/*
		World is the interface used to register components, manage entities and components.

	 The World have the limitation of total MaxComponentCount different components
*/
type World interface {
	// NewEntity creates an entity with optional components and return it's ID
	NewEntity(...ComponentID) EntityID
	// RemEntity removes the entity and it's components from the world
	RemEntity(EntityID)
	// IsAlive returns true if the entity is alive in the world
	IsAlive(EntityID) bool
	// AddComponent adds another component to the entity, if the entity is alive.
	// Adding the same component multiple times is ignored.
	AddComponent(EntityID, ComponentID)
	// RemComponent removes a component fom the entity. This function does nothing if
	// the component don't exist in this entity
	RemComponent(EntityID, ComponentID)
	// Component returns the component pointer for this entity.
	Component(EntityID, ComponentID) unsafe.Pointer
	// Register adds a component registry to the world. If the component ID is
	// already in use, this function panics
	Register(ComponentRegistry)
	// Query returns a QueryCursor for the component mask.
	// You can use the helper function MakeComponentMask(...ComponentID) to create the mask.
	// An empty mask returns a query cursor for all entities in the world.
	Query(Mask) QueryCursor
}

type world struct {
	entityPool EntityPool
	factory    ComponentFactory
	archGraph  ArchetypeGraph
}

/*
NewWorld returns an implementation for the World
*/
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
	column := arch.columns[component]
	if column == nil {
		return nil
	}
	return column.Get(uint(row))
}

func (w *world) Register(comp ComponentRegistry) {
	w.factory.Register(comp)
}

func (w *world) Query(mask Mask) QueryCursor {
	return w.archGraph.Query(mask)
}

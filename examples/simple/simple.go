package main

import (
	"github.com/marioolofo/go-gameengine-ecs"
)

// Component IDs
const (
	TransformID ecs.ID = iota
	PhysicsID
)

type Vec2D struct {
	x, y float32
}

type TransformComponent struct {
	position Vec2D
	rotation float32
}

type PhysicsComponent struct {
	linearAccel, velocity Vec2D
	angularAccel, torque  float32
}

func main() {
	// initial configuration to create the world, new components can be
	// added latter with world.RegisterComponents()
	config := []ecs.ComponentConfig{
		{ID: TransformID, Component: TransformComponent{}},
		{ID: PhysicsID, Component: PhysicsComponent{}},
	}

	// NewWorld allocates a world and register the components
	world := ecs.NewWorld(config...)

	// World.NewEntity will add a new entity to this world
	entity := world.NewEntity()
	// World.Assign adds a list of components to the entity
	// If the entity already have the component, the Assign is ignored
	world.Assign(entity, PhysicsID, TransformID)

	// Any component registered on this entity can be retrieved using World.Component()
	// It's safe to keep this reference until the entity or the component is removed
	phys := (*PhysicsComponent)(world.Component(entity, PhysicsID))
	phys.linearAccel = Vec2D{x: 2, y: 1.5}

	// World.NewFilter creates a cache of entities that have the required components
	//
	// This solution is better than using Systems to update the entities because it's possible to
	// iterate over the filters at variable rate inside your own update function, for example,
	// the script for AI don't need to update at same frequency as physics and animations
	//
	// This filter will be automatically updated when entities or components are added/removed to the world
	filter := world.NewFilter(TransformID, PhysicsID)

	dt := float32(1.0 / 60.0)

	// filter.Entities() returns the updated list of entities that have the required components
	for _, entity := range filter.Entities() {
		// get the components for the entity
		phys := (*PhysicsComponent)(world.Component(entity, PhysicsID))
		tr := (*TransformComponent)(world.Component(entity, TransformID))

		phys.velocity.x += phys.linearAccel.x * dt
		phys.velocity.y += phys.linearAccel.y * dt

		tr.position.x += phys.velocity.x * dt
		tr.position.y += phys.velocity.y * dt

		phys.velocity.x *= 0.99
		phys.velocity.y *= 0.99
	}

	// When a filter is no longer needed, just call World.RemFilter() to remove it from the world
	// This is needed as the filters are updated when the world changes
	world.RemFilter(filter)
}

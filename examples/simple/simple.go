package main

import ecs "github.com/marioolofo/go-gameengine-ecs"

// definitions for component ids
const (
	InputComponentID ecs.ComponentID = iota
	PositionComponentID
	SizeComponentID
	ControllableComponentID
	CameraComponentID
)

type Input struct{ xAxis, yAxis, buttons int }
type Position struct{ x, y float32 }
type Size struct{ width, height float32 }
type Controllable struct{}
type Camera struct{ follows ecs.EntityID }

func main() {
	// NewWorld creates the world. See lowlevel.go example for a behind the scenes version
	world := ecs.NewWorld(10)

	// First we need to register the components to be able to use them:
	// NewComponentRegistry[T] simplify the creation of the component's definition struct
	world.Register(ecs.NewComponentRegistry[Position](PositionComponentID))
	world.Register(ecs.NewComponentRegistry[Size](SizeComponentID))
	world.Register(ecs.NewComponentRegistry[Camera](CameraComponentID))

	// Tag components without data are valid components:
	world.Register(ecs.NewComponentRegistry[Controllable](ControllableComponentID))

	// We can have singleton components. It'll be created when the first entity needs it and
	// every call to this component will always return the same pointer
	world.Register(ecs.NewSingletonComponentRegistry[Input](InputComponentID))

	// NewEntity adds the entity with declared components to the world and return the EntityID.
	gameWorldID := world.NewEntity(InputComponentID)
	playerID := world.NewEntity(ControllableComponentID, PositionComponentID, SizeComponentID)
	cameraID := world.NewEntity(CameraComponentID, PositionComponentID, SizeComponentID)

	// world.Component return the pointer to the entity's component
	cam := (*Camera)(world.Component(cameraID, CameraComponentID))
	cam.follows = playerID
	size := (*Size)(world.Component(cameraID, SizeComponentID))
	size.width = 1920
	size.height = 1080

	// Get the input singleton to update every controllable entity
	actualInput := (*Input)(world.Component(gameWorldID, InputComponentID))

	// World.Query creates a iterator for entities that have the requested components
	//
	// This solution is better than using Systems to update the entities because it's up to
	// the programmer to decide at what rates every group of entities updates.
	// Because of Archetypes, the layout of components in memory are sequentially, making
	// the iterations for the most part access the components linearly in memory
	query := world.Query(ecs.MakeComponentMask(ControllableComponentID, PositionComponentID))
	// QueryCursor.Next will return false when we iterate over all found entities
	for query.Next() {
		// QueryCursor.Component will return the component for this entity
		pos := (*Position)(query.Component(PositionComponentID))
		pos.x += float32(actualInput.xAxis)
		pos.y += float32(actualInput.yAxis)
	}

	// Now, update the cameras with the followed entity's position
	query = world.Query(ecs.MakeComponentMask(CameraComponentID))
	for query.Next() {
		cam = (*Camera)(query.Component(CameraComponentID))
		camPos := (*Position)(query.Component(PositionComponentID))

		// only follow if the entity is still alive:
		if world.IsAlive(cam.follows) {
			followPos := (*Position)(world.Component(cam.follows, PositionComponentID))

			camPos.x = followPos.x
			camPos.y = followPos.y
		} else {
			camPos.x = float32(0.0)
			camPos.y = float32(0.0)
		}
	}
}

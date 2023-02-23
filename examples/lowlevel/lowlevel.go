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
	// Instead of using NewWorld, we can create the necessary pieces to manage our game
	// and encapsulate it the way we like:

	// NewEntityPool creates a pool of recyclable entities:
	entityPool := ecs.NewEntityPool(10)
	// NewComponentFactory creates a factory for our components:
	componentFactory := ecs.NewComponentFactory()
	// NewArchetypeGraph creates an archetype manager
	archGraph := ecs.NewArchetypeGraph(componentFactory)

	// First we need to register the components to be able to use them:
	// NewComponentRegistry[T] simplify the creation of the component's definition struct
	componentFactory.Register(ecs.NewComponentRegistry[Position](PositionComponentID))
	componentFactory.Register(ecs.NewComponentRegistry[Size](SizeComponentID))
	componentFactory.Register(ecs.NewComponentRegistry[Camera](CameraComponentID))

	// Tag components without data are valid components:
	componentFactory.Register(ecs.NewComponentRegistry[Controllable](ControllableComponentID))

	// We can have singleton components. It'll be created when the first entity needs it and
	// every call to this component will always return the same pointer
	componentFactory.Register(ecs.NewSingletonComponentRegistry[Input](InputComponentID))

	// We need to get valid IDs from the entity pool and add them to the archetype manager
	// in order to connect the entities with it's component list
	gameWorldID := entityPool.New()
	playerID := entityPool.New()
	cameraID := entityPool.New()

	// Connect the entities with it's component list
	// The ArchetypeGraph creates the archetypes needed and setup the storage for the
	// components using the ComponentFactory
	archGraph.Add(gameWorldID, InputComponentID)
	archGraph.Add(playerID, ControllableComponentID, PositionComponentID, SizeComponentID)
	archGraph.Add(cameraID, CameraComponentID, PositionComponentID, SizeComponentID)

	// To access the entity components, we request the archetype where it's allocated and
	// the row in the buffers where the components are for this entity:
	arch, row := archGraph.Get(cameraID)

	// arch.Component return the pointer to the component's data in the specific row
	cam := (*Camera)(arch.Component(CameraComponentID, row))
	cam.follows = playerID
	size := (*Size)(arch.Component(SizeComponentID, row))
	size.width = 1920
	size.height = 1080

	// Update every controllable object with the input singleton:
	arch, row = archGraph.Get(gameWorldID)
	actualInput := (*Input)(arch.Component(InputComponentID, row))

	// ArchetypeGraph.Query creates a iterator for entities that have the requested components
	//
	// This solution is better than using Systems to update the entities because it's up to
	// the programmer to decide at what rates every group of entities updates.
	// Because of Archetypes, the layout of components in memory are sequentially, making
	// the iterations for the most part access the components linearly in memory
	query := archGraph.Query(ecs.MakeComponentMask(ControllableComponentID, PositionComponentID))
	// QueryCursor.Next will return false when we iterate over all found entities
	for query.Next() {
		// QueryCursor.Component will return the component for this entity
		pos := (*Position)(query.Component(PositionComponentID))
		pos.x += float32(actualInput.xAxis)
		pos.y += float32(actualInput.yAxis)
	}

	// Now, update the cameras with the followed entity's position
	query = archGraph.Query(ecs.MakeComponentMask(CameraComponentID))
	for query.Next() {
		cam = (*Camera)(query.Component(CameraComponentID))
		camPos := (*Position)(query.Component(PositionComponentID))

		// only follow if the entity is still alive:
		if entityPool.IsAlive(cam.follows) {
			arch, row := archGraph.Get(cam.follows)
			followPos := (*Position)(arch.Component(PositionComponentID, row))

			camPos.x = followPos.x
			camPos.y = followPos.y
		} else {
			camPos.x = float32(0.0)
			camPos.y = float32(0.0)
		}
	}
}

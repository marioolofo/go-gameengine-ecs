package ecs

func test() {
	world := NewWorld()
	// number of threads to execute the loop of searches
	world.SetParallel(4)

	// Entity is a suggar wrapper for the EntityID, with helper functions
	// to add, remove and create relations with other entities.
	camera := world.NewEntity()
	player := world.NewEntity()
	level := world.NewEntity()
	inventory := world.NewEntity()

	// Add adds the components to the entity, and register the component if its not already in the list
	level.Add(NewTileSet("tileset.map1"))
		.Add(&TileSetCollider{})

	camera.Add(&Camera2D{ zoom: 40, fovY: 40.0 / 25.0  })
		.Add(&Follow{ entity: player })
		.ChildOf(level.EntityID)	

	// Next step is to create something like "batch add" to reduce the stress on archetypes
	player.Add(&PlayerControlled{})
		.Add(&Vec3{ 0, 0, 0 })
		.Add(&Rotation{ 0.0 })
		.Add(&BoxCollider{})
		.Add(&Energy{ total: 100, actual: 100 })
		.Add(&Mana{ total: 0, actual: 0 })
		.Add(NewAnimation("player.anim"))
		.Add(&AnimationPlayer{ state: "idle" })
		.ChildOf(level.EntityID)	

	inventory.ChildOf(e.EntityID)
		.Add(&Inventory{})

	allControlledEntities = []EntityID{
		w.ComponentID(&Vec3{})
		w.ComponentID(&Rotation{}),
		w.ComponentID(&PlayerControlled{}),
	}

	allAnimatedEntities = []EntityID {
		w.ComponentID(&AnimationStateMachine{})
		w.ComponentID(&AnimationPlayer{})
	}

	w.Each(allControlledEntities, func (entityID EntityID, vec, rot, _ unsafe.Pointer) {
		//
	})
	// EachParallel runs async and subdivides the calls in batches to speed up the proccess
	w.EachParallel(allAnimatedEntities, func(entityID EntityID, animFSM, animPlayer unsafe.Pointer) {
		//
	})
	w.Each(allFollowCameras, func(entityID EntityID, camView, follow unsafe.Pointer) {
		//
	})
}
package ecs

import (
	"testing"
)

func TestWorldUpdateCycle(t *testing.T) {
	config := []ComponentConfig{
		{Transform2DComponentID, 0, Transform2D{}},
		{Physics2DComponentID, 0, Physics2D{}},
		{ScriptComponentID, 0, Script{}},
	}

	world := NewWorld(config...)

	filter := world.NewFilter(Transform2DComponentID)

	e1 := world.NewEntity()
	world.Assign(e1, Transform2DComponentID)
	e2 := world.NewEntity()
	world.Assign(e2, Transform2DComponentID, Physics2DComponentID)
	e1 = world.NewEntity()
	world.Assign(e1, Transform2DComponentID, ScriptComponentID)

	if filter.World() != world {
		t.Error("filter.World() is different from world!")
	}

	if len(filter.Entities()) != 3 {
		t.Errorf("world.filter with invalid entitie len (expected 3, got %d)\n", len(filter.Entities()))
	}

	for _, entity := range filter.Entities() {
		if world.Component(entity, Transform2DComponentID) == nil {
			t.Errorf("update for group received invalid entity (e %d)\n", entity)
		}
	}

	filter2 := world.NewFilter(ScriptComponentID)

	world.RemEntity(e2)
	e2 = world.NewEntity()
	world.Assign(e2, ScriptComponentID)

	world.RemFilter(filter)

	if len(filter2.Entities()) != 2 {
		t.Errorf("world.filter with invalid entitie len (expected 1, got %d)\n", len(filter.Entities()))
	}

	for _, entity := range filter2.Entities() {
		if world.Component(entity, ScriptComponentID) == nil {
			t.Errorf("update for group received invalid entity (e %d)\n", entity)
		}
	}
}

func TestWorldLockUnlockBatch(t *testing.T) {
	entityCount := 10

	config := []ComponentConfig{
		{Transform2DComponentID, 0, Transform2D{}},
		{Physics2DComponentID, 0, Physics2D{}},
	}

	world := NewWorld(config...)

	entitiesMask := make([]Mask, entityCount)

	for i := 0; i < entityCount; i++ {
		mask := Mask(0)
		mask.Set(Transform2DComponentID, true)
		entitiesMask[i] = mask
	}

	transformFilter := world.NewFilter(Transform2DComponentID)
	physicsFilter := world.NewFilter(Physics2DComponentID)

	entityTransformCount := 0
	entityPhysicsCount := 0

	world.Lock()

	for i := 0; i < len(entitiesMask); i++ {
		entity := world.NewEntity()

		if entitiesMask[i].Get(Transform2DComponentID) {
			world.Assign(entity, Transform2DComponentID)
			entityTransformCount++
		}
		if entitiesMask[i].Get(Physics2DComponentID) {
			world.Assign(entity, Physics2DComponentID)
			entityPhysicsCount++
		}
	}

	world.Unlock()

	if len(transformFilter.Entities()) != entityTransformCount {
		t.Errorf("Added %d entities for transform inside Lock(), filter received %d!\n", entityTransformCount, len(transformFilter.Entities()))
	}
	if len(physicsFilter.Entities()) != entityPhysicsCount {
		t.Errorf("Added %d entities for physics inside Lock(), filter received %d!\n", entityPhysicsCount, len(physicsFilter.Entities()))
	}

	world.RemFilter(transformFilter)
	world.RemFilter(physicsFilter)

	transformFilter = world.NewFilter(Transform2DComponentID)
	physicsFilter = world.NewFilter(Physics2DComponentID)

	if len(transformFilter.Entities()) != entityTransformCount {
		t.Errorf("Entities not added to filter!\n")
	}

	if len(physicsFilter.Entities()) != entityPhysicsCount {
		t.Errorf("Entities not added to filter!\n")
	}

	world.Lock()

	entityTransformCount = 0
	entityPhysicsCount = 0

	for i := 0; i < len(entitiesMask); i++ {
		mask := NewMask(Physics2DComponentID, Transform2DComponentID)
		if entitiesMask[i] == mask {
			world.Remove(Entity(i + 1), Transform2DComponentID, Physics2DComponentID)
			entitiesMask[i] = Mask(0)
			continue
		}

		if entitiesMask[i].Get(Transform2DComponentID) {
			world.Remove(Entity(i + 1), Transform2DComponentID)
			world.Assign(Entity(i + 1), Physics2DComponentID)
			entityPhysicsCount++
		} else {
			world.Remove(Entity(i + 1), Physics2DComponentID)
			world.Assign(Entity(i + 1), Transform2DComponentID)
			entityTransformCount++
		}
		trState := entitiesMask[i].Get(Transform2DComponentID)
		phState := entitiesMask[i].Get(Physics2DComponentID)
		entitiesMask[i].Set(Transform2DComponentID, !trState)
		entitiesMask[i].Set(Physics2DComponentID, !phState)
	}

	world.Unlock()
	world.Unlock()
	world.Unlock()

	if len(transformFilter.Entities()) != entityTransformCount {
		t.Errorf("Added %d entities for transform inside Lock(), filter received %d!\n", entityTransformCount, len(transformFilter.Entities()))
	}
	if len(physicsFilter.Entities()) != entityPhysicsCount {
		t.Errorf("Added %d entities for physics inside Lock(), filter received %d!\n", entityPhysicsCount, len(physicsFilter.Entities()))
	}

	world.Lock()

	if len(transformFilter.Entities()) != entityTransformCount {
		t.Errorf("Added %d entities for transform inside Lock(), filter received %d!\n", entityTransformCount, len(transformFilter.Entities()))
	}
	if len(physicsFilter.Entities()) != entityPhysicsCount {
		t.Errorf("Added %d entities for physics inside Lock(), filter received %d!\n", entityPhysicsCount, len(physicsFilter.Entities()))
	}

	world.Unlock()

	world.Lock()

	for i := 0; i < len(entitiesMask); i++ {
		world.RemEntity(Entity(i + 1))
	}

	world.RemEntity(Entity(1))

	world.Unlock()

	if len(transformFilter.Entities()) != 0 {
		t.Errorf("Entities removed inside Lock(), but transfomFilter has %d entries!\n", len(transformFilter.Entities()))
	}
	if len(physicsFilter.Entities()) != 0 {
		t.Errorf("Entities removed inside Lock(), but physicsFilter has %d entries!\n", len(physicsFilter.Entities()))
	}
}

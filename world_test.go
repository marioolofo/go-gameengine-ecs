package ecs

import (
	"math/rand"
	"testing"
)

const (
	UIDesignComponentID ID = iota
	Transform2DComponentID
	Physics2DComponentID
	Transform3DComponentID
	Physics3DComponentID
	ScriptComponentID
	CustomComponentStartID
)

type UIDesign struct {
	name   string
	config map[string]string
}

type Transform2D struct {
	x, y        float32
	orientation [4]float32
}

type Physics2D struct {
	velocity, torque float32
}

type Script struct {
	filePath string
}

func TestWorldCore(t *testing.T) {
	config := []ComponentConfig{
		{UIDesignComponentID, 0, UIDesign{}},
		{Transform2DComponentID, 0, Transform2D{}},
		{Physics2DComponentID, 0, Physics2D{}},
		{ScriptComponentID, 0, Script{}},
		{ScriptComponentID, 0, Script{}},
		{999, 0, Script{}},
	}

	world := NewWorld(config...)
	entities := make([]Entity, 100000, 100000)

	for i := 1; i < 100000; i++ {
		entity := world.NewEntity()
		world.Assign(entity, Transform2DComponentID)

		var ids []ID

		if rand.Float32() < 0.5 {
			ids = append(ids, UIDesignComponentID)
		}
		if rand.Float32() < 0.5 {
			ids = append(ids, Physics2DComponentID)
		}
		if rand.Float32() < 0.5 {
			ids = append(ids, ScriptComponentID)
		}

		if rand.Float32() < 0.1 {
			ids = append(ids, 999)
		}
		if rand.Float32() < 0.1 {
			ids = append(ids, MaskTotalBits-1)
		}

		world.Assign(entity, ids...)
		entities[i] = entity

		if entity != ID(i) {
			t.Errorf("entity id %d is different from expected %d\n", entity, i)
		}

		tr := (*Transform2D)(world.Component(entity, Transform2DComponentID))
		if tr != nil {
			tr.x = float32(i)
			tr.y = float32(i * 2)
			tr.orientation[2] = float32(i * 3)
		}
	}

	for i := 1; i < 100000; i++ {
		entity := entities[i]

		tr := (*Transform2D)(world.Component(entity, Transform2DComponentID))
		if tr != nil {
			diff := tr.x + tr.y - float32(i*3)
			if diff < -0.0001 || diff > 0.0001 {
				t.Errorf("entity %d don't have the correct values for component Transform2D (received %v)\n", i, tr)
			}
		}

		comp := world.Component(entity, ID(32))

		if comp != nil {
			t.Errorf("entity received valid pointer for invalid component\n")
		}

		world.Remove(entity, UIDesignComponentID)
		world.RemEntity(entity)
	}
	newEntity := world.NewEntity()
	// IDs are recycled from last to first entity removed from the world
	if newEntity != Entity(99999) {
		t.Errorf("expected entity recycle for id 1, received %d\n", newEntity)
	}
}

func TestWorldEntityID(t *testing.T) {
	config := []ComponentConfig{
		{Transform2DComponentID, 0, Transform2D{}},
	}
	world := NewWorld(config...)

	entity := world.NewEntity()
	if entity < 1 {
		t.Error("entity ID starting at wrong index:", entity)
	}

	world.Assign(0, Transform2DComponentID)
	world.Remove(0, Transform2DComponentID)

	invalidComponentPtr := world.Component(0, Transform2DComponentID)
	if invalidComponentPtr != nil {
		t.Error("expected nil, received valid component for invalid entity")
	}

	filter := world.NewFilter()

	if len(filter.Entities()) != 1 {
		t.Errorf("expected filter with 1 entity, received %d", len(filter.Entities()))
	}

	world.RemEntity(entity)

	if len(filter.Entities()) != 0 {
		t.Errorf("expected filter with no entity, received %d", len(filter.Entities()))
	}

	world.RemEntity(0)

	entity = world.NewEntity()
	entity2 := world.NewEntity()
	if entity != 1 || entity2 != 2 {
		t.Errorf("expected entities IDs 1 & 2, received %d & %d", entity, entity2)
	}
}

func TestWorldDebug(t *testing.T) {
	config := []ComponentConfig{
		{Transform2DComponentID, 0, Transform2D{}},
	}
	world := NewWorld(config...)

	entity := world.NewEntity()
	world.RemEntity(entity)

	world.Assign(entity, Transform2DComponentID)
	world.Remove(entity, Transform2DComponentID)
	world.RemEntity(entity)
}

func TestWorldStats(t *testing.T) {
	config := []ComponentConfig{
		{UIDesignComponentID, 0, UIDesign{}},
		{Transform2DComponentID, 0, Transform2D{}},
		{Physics2DComponentID, 0, Physics2D{}},
		{ScriptComponentID, 0, Script{}},
	}

	world := NewWorld(config...)

	stats := world.Stats()

	if (stats.EntityStats.Recycled != 0) {
		t.Errorf("expected to have 0 entities recycled, found %d\n", stats.EntityStats.Recycled)
	}
	if (stats.EntityStats.InUse != 0) {
		t.Errorf("expected to have 0 entities, found %d\n", stats.EntityStats.InUse)
	}
	if (stats.EntityStats.Total != 0) {
		t.Errorf("expected to have a total of 0 entities, found %d\n", stats.EntityStats.Total)
	}

	if (stats.ComponentCount != uint(len(config))) {
		t.Errorf("expected to have a total of %d components, found %d\n", len(config), stats.ComponentCount)
	}

	world.NewEntity()
	entity := world.NewEntity()
	world.NewEntity()

	stats = world.Stats()

	if (stats.EntityStats.Recycled != 0) {
		t.Errorf("expected to have 0 entities recycled, found %d\n", stats.EntityStats.Recycled)
	}
	if (stats.EntityStats.InUse != 3) {
		t.Errorf("expected to have 3 entities in use, found %d\n", stats.EntityStats.InUse)
	}
	if (stats.EntityStats.Total != 3) {
		t.Errorf("expected to have a total of 3 entities, found %d\n", stats.EntityStats.Total)
	}

	world.RemEntity(entity)

	stats = world.Stats()

	if (stats.EntityStats.Recycled != 1) {
		t.Errorf("expected to have 1 entities recycled, found %d\n", stats.EntityStats.Recycled)
	}
	if (stats.EntityStats.InUse != 2) {
		t.Errorf("expected to have 2 entities in use, found %d\n", stats.EntityStats.InUse)
	}
	if (stats.EntityStats.Total != 3) {
		t.Errorf("expected to have a total of 3 entities, found %d\n", stats.EntityStats.Total)
	}

	invalidStats := world.ComponentStats(9999)
	accum := invalidStats.SparseArrayLength + invalidStats.MemStats.Alignment + invalidStats.MemStats.InUse

	if (accum != 0) {
		t.Errorf("invalid component returned unknown status %v\n", invalidStats)
	}
}

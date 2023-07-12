package ecs

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestWorld(t *testing.T) {
	const Vec3CompID ComponentID = 0
	type Vec3 struct{}

	world := NewWorld(100)
	assert.NotNil(t, world, "NewWorld should return a valid World object")

	world.Register(NewComponentRegistry[Vec3](Vec3CompID))

	e1 := world.NewEntity()
	world.AddComponent(e1, Vec3CompID)
	world.RemComponent(e1, Vec3CompID)
	assert.NotZero(t, e1.ID(), "NewEntity should return a valid entity > 0")
	assert.True(t, world.IsAlive(e1), "expected entity to be alive")

	e2 := world.NewEntity(Vec3CompID)
	assert.NotZero(t, e2.ID(), "NewEntity with components should return a valid entity")
	assert.True(t, world.Component(e2, Vec3CompID) != unsafe.Pointer(nil), "Component() should return a valid pointer for valid components")

	query := world.Query(MakeComponentMask(Vec3CompID))
	assert.True(t, query.Next(), "Query should return a valid cursor")
	assert.True(t, query.Entity() == e2, "expected Query.Entity to be %x, received %x", e2, query.Entity())
	assert.True(t, query.Component(Vec3CompID) != unsafe.Pointer(nil), "query.Component() should return a valid pointer for valid components")

	world.RemComponent(e2, Vec3CompID)

	query = world.Query(Mask{})
	count := 0
	for query.Next() {
		count++
	}
	assert.True(t, count == 2, "expected query with empty mask to return all entities")

	world.RemEntity(e1)
	assert.False(t, world.IsAlive(e1), "expected IsAlive to return false for invalid entities")
}

func Test_ECS_Component(t *testing.T) {
	const (
		CompAID ComponentID = iota
		CompBID
	)
	type CompA struct{}
	type CompB struct{}

	w := NewWorld(0)

	w.Register(NewComponentRegistry[CompA](CompAID))
	w.Register(NewComponentRegistry[CompB](CompBID))

	someEntity := w.NewEntity(CompAID)

	b := (*CompB)(w.Component(someEntity, CompBID))
	assert.Nil(t, b)
}

package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentFactory(t *testing.T) {
	const (
		Vec3CompID = iota
		AmnoCompID
		ConfigCompID
	)
	type Vec3 struct{ x, y, z float32 }
	type Amno struct{ quantity int }
	type Config struct {
		playerCount uint
		score       []uint
	}

	factory := NewComponentFactory()

	vec3Comp := NewComponentRegistry[Vec3](Vec3CompID)
	amnoComp := NewComponentRegistry[Amno](AmnoCompID)

	assert.NotNil(t, vec3Comp, "NewComponentRegistry[Vec3] should return a Component")
	assert.NotNil(t, amnoComp, "NewComponentRegistry[Amno] should return a Component")

	factory.Register(vec3Comp)
	factory.Register(amnoComp)

	comp, ok := factory.GetByType(&Vec3{})
	assert.True(t, ok, "GetByType(&Vec3{}) should return Component ref")
	assert.NotNil(t, comp, "GetByType(&Vec3{}) should return Component ref")
	assert.True(t, vec3Comp.id == comp.id, "GetByType(&Vec3{}) should return the corect Component ref")

	comp, ok = factory.GetByID(AmnoCompID)
	assert.True(t, ok, "GetByID(&Amno{}) should return Component ref")
	assert.NotNil(t, comp, "GetByID(&Amno{}) should return Component ref")
	assert.True(t, amnoComp.id == comp.id, "GetByID(&Amno{}) should return the corect Component ref")

	storage := vec3Comp.NewStorage()
	assert.NotNil(t, storage, "comp.NewStorage() should return a valid Storage")
}

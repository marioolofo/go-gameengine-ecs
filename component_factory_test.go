package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentFactory(t *testing.T) {
	type Vec3 struct{ x, y, z float32 }
	type Amno struct{ quantity int }
	type Config struct {
		playerCount uint
		score       []uint
	}

	ep := NewEntityPool(0)
	factory := NewComponentFactory()

	vec3Comp := NewComponentRegistry[Vec3](ep)
	amnoComp := NewComponentRegistry[Amno](ep)
	configComp := NewComponentSingletonRegistry[Config](ep)

	assert.NotNil(t, vec3Comp, "NewComponentRegistry[Vec3] should return a Component")
	assert.NotNil(t, amnoComp, "NewComponentRegistry[Amno] should return a Component")
	assert.NotNil(t, configComp, "NewComponentSingletonRegistry[Config] should return a Component")

	id := factory.Register(vec3Comp)
	// assert.True(t, id.IsComponent(), "Register() should return valid id for new Component")

	id = factory.Register(amnoComp)
	// assert.True(t, id.IsComponent(), "Register() should return valid id for new Component")

	id = factory.Register(configComp)
	// assert.True(t, id.IsComponent(), "Register() should return valid id for new Component")

	if !id.IsChild() {
		//
	}

	comp, ok := factory.GetByType(Vec3{})
	assert.True(t, ok, "GetByType(&Vec3{}) should return Component ref")
	assert.NotNil(t, comp, "GetByType(&Vec3{}) should return Component ref")
	assert.True(t, vec3Comp == comp, "GetByType(&Vec3{}) should return the corect Component ref")

	comp, ok = factory.GetByID(amnoComp.EntityID)
	assert.True(t, ok, "GetByID(&Amno{}) should return Component ref")
	assert.NotNil(t, comp, "GetByID(&Amno{}) should return Component ref")
	assert.True(t, amnoComp == comp, "GetByID(&Amno{}) should return the corect Component ref")

	storage := vec3Comp.NewStorage(0, 0)
	assert.NotNil(t, storage, "comp.NewStorage() should return a valid Storage")

	storage = configComp.NewStorage(0, 0)
	assert.Nil(t, storage, "Singleton component don't have storage")
	assert.NotZero(t, configComp.SingletonPtr, "Singleton component should have a valid struct pointer")
}

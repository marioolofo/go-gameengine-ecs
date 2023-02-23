package ecs

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestComponentFactory(t *testing.T) {
	const (
		Vec3CompID = iota
		AmnoCompID
		ConfigCompID
		InputCompID
	)
	//lint:file-ignore U1000 we are testing different sized structs
	type Vec3 struct{ x, y, z float32 }
	type Amno struct{ quantity int }
	type Config struct {
		playerCount uint
		score       []uint
	}
	type Tag struct{}
	type Input struct {
		name          string
		axis, buttons int
	}

	factory := NewComponentFactory()

	vec3Comp := NewComponentRegistry[Vec3](Vec3CompID)
	amnoComp := NewComponentRegistry[Amno](AmnoCompID)
	configComp := NewComponentRegistry[Config](ConfigCompID)
	singlComp := NewSingletonComponentRegistry[Input](InputCompID)

	assert.NotNil(t, vec3Comp, "NewComponentRegistry[Vec3] should return a Component")
	assert.NotNil(t, amnoComp, "NewComponentRegistry[Amno] should return a Component")

	factory.Register(vec3Comp)
	factory.Register(amnoComp)
	factory.Register(configComp)
	factory.Register(singlComp)

	assert.Panics(t, func() {
		factory.Register(amnoComp)
	}, "should panic when registering same component twice")

	comp, ok := factory.GetByType(&Vec3{})
	assert.True(t, ok, "GetByType(&Vec3{}) should return Component ref")
	assert.NotNil(t, comp, "GetByType(&Vec3{}) should return Component ref")
	assert.True(t, vec3Comp.ID == comp.ID, "GetByType(&Vec3{}) should return the corect Component ref")

	comp, ok = factory.GetByID(AmnoCompID)
	assert.True(t, ok, "GetByID(&Amno{}) should return Component ref")
	assert.NotNil(t, comp, "GetByID(&Amno{}) should return Component ref")
	assert.True(t, amnoComp.ID == comp.ID, "GetByID(&Amno{}) should return the corect Component ref")

	storage := vec3Comp.NewStorage()
	assert.NotNil(t, storage, "comp.NewStorage() should return a valid Storage")

	storage = configComp.NewStorage()
	assert.NotNil(t, storage, "comp.NewStorage() should return a valid Storage")

	v := storage.Get(0)
	assert.False(t, v == unsafe.Pointer(nil), "storage.Get should return valid pointer even for zero sized structs")

	storage = singlComp.NewStorage()
	storage2 := singlComp.NewStorage()
	assert.Equal(t, storage, storage2, "singleton storage should always return the same Storage")
	ptr := storage.Get(0)
	ptr2 := storage.Get(1000)
	assert.Equal(t, ptr, ptr2, "singleton storage should return the same address for any index")

	input := Input{"Keyboard", 0xacacacac, 0xf0f0f0f0}
	storage.Set(0, &input)

	inputPtr := (*Input)(ptr)
	assert.EqualValues(t, *inputPtr, input, "singletonStorage.Set should set the value correctly")

	input.axis = 0x55005500
	input.buttons = 0xb3b3b3b3

	storage.Copy(0, unsafe.Pointer(&input))
	assert.EqualValues(t, *inputPtr, input, "singletonStorage.Copy should set the value correctly")

	input = Input{}
	storage.Reset()
	assert.EqualValues(t, *inputPtr, input, "singletonStorage.Reset should clear the value correctly")

	stats := storage.Stats()
	assert.EqualValues(t, stats.Type, reflect.TypeOf(input), "singletonStorage.Stats should return the correct typeOf")
	assert.True(t, stats.Cap == 1, "singletonStorage.Cap should be set to one")
	assert.True(t, uintptr(stats.ItemSize) == unsafe.Sizeof(input), "singletonStorage.ItemSize should be the size of the struct")
}

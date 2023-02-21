package ecs

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestStorageAllocation(t *testing.T) {
	type vec3 struct{ x, y, z float32 }
	var v vec3

	testStorageAllocation := func(t *testing.T, s Storage) {
		stats := s.Stats()

		assert.EqualValues(t, unsafe.Sizeof(v), stats.ItemSize, "struct size mismatch: got %d, want %d", stats.ItemSize, unsafe.Sizeof(v))
		assert.EqualValues(t, stats.Cap, uint(1), "capacity mismatch: got %d, want %d", stats.Cap, 1)
		assert.Equal(t, stats.Type, reflect.TypeOf(v), "type mismatch: got %+v, want %+v", stats.Type, reflect.TypeOf(v))

		s.Expand(4)
		vp := (*vec3)(s.Get(3))

		assert.NotNil(t, vp, "expected vec3 at index 3, got nil instead")

		v.x = 1
		v.y = 2
		v.z = 3

		result := s.Set(10, &v)
		assert.False(t, result, "expected Set() to return false for invalid index")

		s.Expand(11)
		v10 := (*vec3)(s.Get(10))
		assert.NotNil(t, v10, "expected vec3 at index 10, got nil instead")

		result = s.Set(10, &v)
		assert.True(t, result, "expected Set() to return true for valid index")

		assert.Equal(t, v10.x, float32(1.0), "expected v10.x to be 1.0, got %f instead", v10.x)
		assert.Equal(t, v10.y, float32(2.0), "expected v10.y to be 2.0, got %f instead", v10.y)
		assert.Equal(t, v10.z, float32(3.0), "expected v10.z to be 3.0, got %f instead", v10.z)

		s.Reset()
		stats = s.Stats()
		assert.EqualValues(t, stats.Cap, uint(0), "capacity mismatch: got %d, want %d", stats.Cap, 0)
	}

	genericStorage := NewStorage[vec3](1, 5)

	testStorageAllocation(t, genericStorage)
}

func TestStorageRemove(t *testing.T) {
	type vec3 struct{ x, y, z float32 }

	s := NewStorage[vec3](10, 0)

	for i := 0; i < 10; i++ {
		result := s.Set(uint(i), &vec3{float32(i), float32(i), float32(i)})
		assert.True(t, result, "expected Set() to return true for valid index")
	}
	expected := []float32{0, 1, 2, 3, 4}
	s.Shrink(uint(5))
	for i, f := range expected {
		v := (*vec3)(s.Get(uint(i)))
		assert.Equal(t, v, &vec3{f, f, f}, "expected vec3 to be %+v, got %+v", vec3{f, f, f}, v)
	}
}

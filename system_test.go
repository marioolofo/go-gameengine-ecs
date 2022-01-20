package ecs

import "testing"

type vec3 struct {
	x, y, z float32
}

func TestSystemNew(t *testing.T) {
	sys := NewSystem(1, vec3{}, -1)

	v0 := (*vec3)(sys.New(0))
	if v0 == nil {
		t.Fatal("System.New() returned nil pointer")
	}

	v0eq := (*vec3)(sys.New(0))
	if v0 != v0eq {
		t.Fatal("System.New() invalid pointer for ID already reserved")
	}
	v0eq = (*vec3)(sys.Get(0))
	if v0 != v0eq {
		t.Fatal("System.Get() returned different pointer from New()")
	}

	v10 := sys.New(10)
	if v10 == nil {
		t.Fatal("System.New() returned nil pointer")
	}

	v2 := (*vec3)(sys.Get(2))
	if v2 != nil {
		t.Fatal("System.Get() valid pointer for invalid index")
	}

	sys.Recycle(0)

	v0eq = (*vec3)(sys.Get(0))
	if v0eq != nil {
		t.Fatal("recycled index returned data")
	}

	sys.Reset()

	v0eq = (*vec3)(sys.Get(0))
	if v0eq != nil {
		t.Fatal("system still contains valid data after Reset()")
	}
}

func TestSystemSetAndZero(t* testing.T) {
	sys := NewSystem(1, vec3{}, 0)

	iterations := 100000
	iterOffset := 3

	for i := 0; i < iterations; i += iterOffset {
		v := (*vec3)(sys.New(ID(i)))
		v.x = float32(i)
		v.y = float32(i)
		v.z = float32(i)

		sys.Set(ID(i), v)

		vv := (*vec3)(sys.New(ID(i)))
		if v != vv {
			t.Error("expected already added ID to return previous instance")
		}
	}

	for i := 0; i < iterations; i++ {
		v := (*vec3)(sys.Get(ID(i)))
		if i % iterOffset != 0 && v != nil {
			t.Errorf("invalid index %d contains valid object!", i)
		} else if i % iterOffset == 0 {
			if v == nil {
				t.Errorf("valid index %d contains null object!", i)
			} else {
				ind := float32(i)
				err := v.x-ind + v.y-ind + v.z-ind
				if err > 0.00001 || err < -0.00001 {
					t.Errorf("vector retrieved with invalid values! (%f, %f, %f)\n", v.x, v.y, v.z)
				}
				sys.Zero(ID(i))
				if v.x != 0 || v.y != 0 || v.z != 0 {
					t.Errorf("struct still contains data after Zero() (%+v)\n", v)
				}
			}
		}
	}
}

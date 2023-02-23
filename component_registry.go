package ecs

import (
	"reflect"
	"unsafe"
)

type ComponentRegistry struct {
	ID ComponentID
	reflect.Type
	NewStorage func() Storage
}

func MakeComponentMask(bits ...ComponentID) Mask {
	mask := Mask{}
	for _, bit := range bits {
		mask.Set(uint64(bit))
	}
	return mask
}

func NewComponentRegistry[T any](id ComponentID) ComponentRegistry {
	var t T
	typeOf := reflect.TypeOf(t)

	return ComponentRegistry{
		id,
		typeOf,
		func() Storage {
			return NewStorage[T](ComponentStorageInitialCap, ComponentStorageIncrement)
		},
	}
}

func NewSingletonComponentRegistry[T any](id ComponentID) ComponentRegistry {
	var t T
	typeOf := reflect.TypeOf(t)

	return ComponentRegistry{
		id,
		typeOf,
		func() Storage {
			return newSingletonStorage[T]()
		},
	}
}

type singletonStorage[T any] struct {
	value T
}

func newSingletonStorage[T any]() Storage {
	return &singletonStorage[T]{}
}

func (s *singletonStorage[T]) Get(uint) unsafe.Pointer {
	return unsafe.Pointer(&s.value)
}

func (s *singletonStorage[T]) Set(pos uint, value interface{}) bool {
	ptr, ok := value.(*T)
	if ok {
		s.value = *ptr
	}
	return ok
}

func (s *singletonStorage[T]) Copy(pos uint, ptr unsafe.Pointer) {
	val := (*T)(ptr)
	s.value = *val
}

func (s *singletonStorage[T]) Shrink(uint) {}

func (s *singletonStorage[T]) Expand(uint) {}

func (s *singletonStorage[T]) Reset() {
	var t T
	s.value = t
}

func (s singletonStorage[T]) Stats() StorageStats {
	var t T
	typeOf := reflect.TypeOf(t)

	return StorageStats{
		typeOf,
		uint(typeOf.Size()),
		1,
	}
}

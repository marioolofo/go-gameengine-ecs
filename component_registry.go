package ecs

import (
	"reflect"
	"unsafe"
)

/*
ComponentRegistry defines a component ID, it's type and how to create a new Storage for it.
*/
type ComponentRegistry struct {
	// ID defines the identifier for this component
	ID ComponentID
	// Type is the type of the component
	reflect.Type
	// NewStorage is a factory function that returns an implementation of the Storage interface
	NewStorage func() Storage
	singleton  Storage
}

// NewComponentRegistry[T] returns a ComponentRegistry definition for the type T and id
func NewComponentRegistry[T any](id ComponentID) ComponentRegistry {
	var t T
	typeOf := reflect.TypeOf(t)

	return ComponentRegistry{
		id,
		typeOf,
		func() Storage {
			return NewStorage[T](ComponentStorageInitialCap, ComponentStorageIncrement)
		},
		nil,
	}
}

// NewSingletonComponentRegistry[T] returns a ComponentRegistry definition for the type T and id,
// with the difference that the NewStorage always returns the same Storage for every call.
func NewSingletonComponentRegistry[T any](id ComponentID) ComponentRegistry {
	var t T
	typeOf := reflect.TypeOf(t)

	storage := newSingletonStorage[T]()

	return ComponentRegistry{
		id,
		typeOf,
		func() Storage {
			return storage
		},
		storage,
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

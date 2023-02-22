package ecs

import (
	"reflect"
	"unsafe"
)

const (
	MaxComponentCount          = 256
	ComponentStorageInitialCap = 1024
	ComponentStorageIncrement  = 2048
)

type ComponentID = uint

func MakeComponentMask(bits ...ComponentID) Mask {
	mask := Mask{}
	for _, bit := range bits {
		mask.Set(uint64(bit))
	}
	return mask
}

type ComponentRegistry struct {
	id ComponentID
	reflect.Value
	SingletonPtr unsafe.Pointer
	NewStorage   func() Storage
}

type ComponentFactory struct {
	refs       map[reflect.Type]uint
	components [MaxComponentCount]ComponentRegistry
}

func NewComponentRegistry[T any](id ComponentID) ComponentRegistry {
	var t T
	typeOf := reflect.TypeOf(t)
	value := reflect.New(typeOf)

	return ComponentRegistry{
		id,
		value,
		value.UnsafePointer(),
		func() Storage {
			return NewStorage[T](ComponentStorageInitialCap, ComponentStorageIncrement)
		},
	}
}

func NewComponentFactory() ComponentFactory {
	return ComponentFactory{
		refs:       make(map[reflect.Type]uint),
		components: [MaxComponentCount]ComponentRegistry{},
	}
}

func (c *ComponentFactory) Register(comp ComponentRegistry) {
	if c.components[comp.id].SingletonPtr != unsafe.Pointer(nil) {
		panic("Component already registered")
	}

	c.refs[comp.Value.Type()] = comp.id
	c.components[comp.id] = comp
}

func (c *ComponentFactory) GetByType(typ interface{}) (*ComponentRegistry, bool) {
	t := reflect.TypeOf(typ)
	comp, ok := c.refs[t]

	var reg *ComponentRegistry
	if ok {
		reg = &c.components[comp]
	}

	return reg, ok
}

func (c *ComponentFactory) GetByID(id ComponentID) (*ComponentRegistry, bool) {
	if c.components[id].SingletonPtr == unsafe.Pointer(nil) {
		return nil, false
	}

	return &c.components[id], true
}

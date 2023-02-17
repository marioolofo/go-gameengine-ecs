package ecs

import (
	"reflect"
	"unsafe"
)

type Component struct {
	EntityID
	reflect.Type
	reflect.Value
	SingletonPtr unsafe.Pointer
	NewStorage   func() Storage
}

type ComponentFactory interface {
	Register(comp *Component) EntityID
	GetByType(typ interface{}) (*Component, bool)
	GetByID(id EntityID) (*Component, bool)
}

type componentFactory struct {
	refs map[reflect.Type]*Component
	ids  map[uint64]*Component
}

func NewComponentRegistry[T any](ep EntityPool) *Component {
	return newComponentRegistry[T](ep, false)
}

func NewComponentSingletonRegistry[T any](ep EntityPool) *Component {
	return newComponentRegistry[T](ep, true)
}

func NewComponentFactory() ComponentFactory {
	return &componentFactory{
		make(map[reflect.Type]*Component),
		make(map[uint64]*Component),
	}
}

func (c *componentFactory) Register(comp *Component) EntityID {
	_, ok := c.ids[comp.EntityID.ID()]
	if !ok {
		c.ids[comp.EntityID.ID()] = comp
		c.refs[comp.Type] = comp
	}
	return comp.EntityID
}

func (c *componentFactory) GetByType(typ interface{}) (*Component, bool) {
	t := reflect.TypeOf(typ)
	comp, ok := c.refs[t]

	return comp, ok
}

func (c *componentFactory) GetByID(id EntityID) (*Component, bool) {
	comp, ok := c.ids[id.ID()]
	return comp, ok
}

func newComponentRegistry[T any](ep EntityPool, singleton bool) *Component {
	var t T
	typeOf := reflect.TypeOf(t)

	id := ep.NewComponent()
	if singleton {
		id = id.Singleton()
	}

	value := reflect.New(typeOf)

	return &Component{
		id,
		typeOf,
		value,
		value.UnsafePointer(),
		func() Storage {
			if singleton {
				return nil
			}
			return NewStorage[T](0, 0)
		},
	}
}
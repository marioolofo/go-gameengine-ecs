package ecs

import (
	"reflect"
)

const (
	MaxComponentCount          uint = 256
	ComponentStorageInitialCap uint = 1024
	ComponentStorageIncrement  uint = 2048
)

type ComponentID = uint

type ComponentFactory interface {
	Register(comp ComponentRegistry)
	GetByType(typ interface{}) (*ComponentRegistry, bool)
	GetByID(id ComponentID) (*ComponentRegistry, bool)
}

type componentFactory struct {
	refs       map[reflect.Type]uint
	components [MaxComponentCount]ComponentRegistry
	mask       Mask
}

func NewComponentFactory() ComponentFactory {
	return &componentFactory{
		refs:       make(map[reflect.Type]uint),
		components: [MaxComponentCount]ComponentRegistry{},
		mask:       Mask{},
	}
}

func (c *componentFactory) Register(comp ComponentRegistry) {
	if c.mask.IsSet(uint64(comp.ID)) {
		panic("Component already registered")
	}

	c.refs[comp.Type] = comp.ID
	c.components[comp.ID] = comp
	c.mask.Set(uint64(comp.ID))
}

func (c *componentFactory) GetByType(typ interface{}) (*ComponentRegistry, bool) {
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	comp, ok := c.refs[t]

	var reg *ComponentRegistry
	if ok {
		reg = &c.components[comp]
	}

	return reg, ok
}

func (c *componentFactory) GetByID(id ComponentID) (*ComponentRegistry, bool) {
	if !c.mask.IsSet(uint64(id)) {
		return nil, false
	}

	return &c.components[id], true
}

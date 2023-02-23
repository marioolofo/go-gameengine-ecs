package ecs

import (
	"reflect"
)

const (
	// the maximum number of components that can be stored
	MaxComponentCount uint = 256
	// initial number of elements in the component Storage
	ComponentStorageInitialCap uint = 1024
	// number of elements that must be added to the Storage when needed
	ComponentStorageIncrement uint = 2048
)

/*
	ComponentFactory is the interface used to store and search for component definitions.
	It's implemented this way to make the code modular and provides a way to share a factory
	between multiple worlds/archetype graphs

	Register registers the component definition in this factory

	GetByType returns the ComponentRegistry for the component of a given type

	GetByID returns the ComponentRegistry for the component id
*/
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

// NewComponentFactory returns an implementation of the ComponentFactory interface
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

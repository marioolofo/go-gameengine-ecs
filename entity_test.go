package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityMake(t *testing.T) {
	testCases := []struct {
		desc   string
		entity EntityID
		id     uint64
		gen    uint64
		flags  EntityID
	}{
		{
			"empty entity",
			MakeEntityWithFlags(0, 0, 0),
			0, 0, 0,
		},
		{
			"entity with ID",
			MakeEntityWithFlags(10, 0, 0),
			10, 0, 0,
		},
		{
			"entity with ID and Gen",
			MakeEntityWithFlags(10, 28, 0),
			10, 28, 0,
		},
		{
			"entity with childOf",
			MakeEntityWithFlags(10, 28, FlagEntityChildOf),
			10, 28, FlagEntityChildOf,
		},
		{
			"entity with instanceOf",
			MakeEntityWithFlags(10, 28, FlagEntityInstanceOf),
			10, 28, FlagEntityInstanceOf,
		},
		{
			"entity with childOf and cloneOf",
			MakeEntityWithFlags(10, 28, FlagEntityInstanceOf|FlagEntityChildOf),
			10, 28, FlagEntityInstanceOf | FlagEntityChildOf,
		},
		{
			"entity disabled",
			MakeEntityWithFlags(10, 28, FlagEntityDisabled),
			10, 28, FlagEntityDisabled,
		},
		{
			"component",
			MakeEntityWithFlags(10, 0, FlagEntityComponent),
			10, 0, FlagEntityComponent,
		},
		{
			"singleton",
			MakeEntityWithFlags(10, 0, FlagEntitySingleton),
			10, 0, FlagEntitySingleton,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			id := tC.entity.ID()
			gen := tC.entity.Gen()
			flags := tC.entity.Flags()

			assert.Equal(t, tC.id, id, "expected entity id %d, found %d", tC.id, id)
			assert.Equal(t, tC.gen, gen, "expected entity gen %d, found %d", tC.gen, gen)
			assert.Equal(t, tC.flags, flags, "expected entity flags %0x, found %0x", tC.flags, flags)

			entity := tC.entity.SetID(1234)
			assert.Equal(t, uint64(1234), entity.ID(), "expected entity id %d, found %d", 1234, entity.ID())
			assert.Equal(t, tC.gen, entity.Gen(), "expected entity gen %d, found %d", tC.gen, entity.Gen())
			assert.Equal(t, tC.flags, entity.Flags(), "expected entity flags %0x, found %0x", tC.flags, entity.Flags())

			entity = tC.entity.Component()
			assert.True(t, entity.IsComponent(), "component entity must have the component flag")
			assert.Equal(t, entity.Gen(), uint64(0), "components don't have generation data")

			entity = tC.entity.Singleton()
			assert.True(t, entity.IsSingleton(), "singleton should have the flag set")
			assert.True(t, entity.IsComponent(), "singleton should have the component flag set")
		})
	}
}

func TestEntityEnableDisable(t *testing.T) {
	e := MakeEntityWithFlags(0, 0, 0)

	e = e.Disable()
	assert.Equal(t, e.Flags(), FlagEntityDisabled, "expected disabled flag")
	assert.Equal(t, e.IsDisabled(), true, "e.Disable() should disable the entity")

	e = e.Enable()
	assert.Equal(t, e.Flags(), EntityID(0), "expected zero flag")
	assert.Equal(t, e.IsDisabled(), false, "e.Enable() should enable the entity")
}

func TestEntityChildOf(t *testing.T) {
	e := MakeEntityWithFlags(0, 0, 0)

	e = e.ChildOf(true)
	assert.Equal(t, e.Flags(), FlagEntityChildOf, "expected childOf flag")
	assert.Equal(t, e.IsChild(), true, "e.IsChild() should be true")

	e = e.ChildOf(false)
	assert.Equal(t, e.Flags(), EntityID(0), "expected zero flag")
	assert.Equal(t, e.IsChild(), false, "e.IsChild() should be false")
}

func TestEntityInstanceOf(t *testing.T) {
	e := MakeEntityWithFlags(0, 0, 0)

	e = e.InstanceOf(true)
	assert.Equal(t, e.Flags(), FlagEntityInstanceOf, "expected instanceOf flag")
	assert.Equal(t, e.IsInstance(), true, "e.IsInstance() should be true")

	e = e.InstanceOf(false)
	assert.Equal(t, e.Flags(), EntityID(0), "expected zero flag")
	assert.Equal(t, e.IsInstance(), false, "e.IsInstance() should be false")
}

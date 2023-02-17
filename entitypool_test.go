package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityPoolAllocation(t *testing.T) {
	ep := NewEntityPool(10)

	entity := ep.New()

	assert.EqualValues(t, entity.ID(), 1, "expected New() to return 1, got %d", entity.ID())
	assert.EqualValues(t, entity.Gen(), 0, "expected New() to return 0 gen, got %d", entity.Gen())

	assert.True(t, ep.IsAlive(entity), "expected IsAlive() to return true for valid entity")
	assert.True(t, ep.Recycle(entity), "expected Recycle() to return true for valid entity")

	assert.False(t, ep.IsAlive(entity), "expected IsAlive() to return false for recycled entity")
	assert.False(t, ep.Recycle(entity), "expected Recycle() to return false for recycled entity")
}

func TestEntityPoolIsAlive(t *testing.T) {
	ep := NewEntityPool(2)

	entity := ep.New()

	assert.True(t, ep.IsAlive(entity), "expected IsAlive() to return true for valid entity")
	assert.True(t, ep.IsAlive(entity.Disable()), "expected IsAlive() to return true for valid entity")
	assert.True(t, ep.IsAlive(entity.InstanceOf(true)), "expected IsAlive() to return true for valid entity")
	assert.False(t, ep.IsAlive(entity.SetID(1000)), "expected IsAlive() to return false for invalid entity (entity %x)", entity.SetID(1000))
}

func TestEntityPoolRecycle(t *testing.T) {
	iter := uint(1000)

	ep := NewEntityPool(0)

	entities := make([]EntityID, 0)

	for i := uint(0); i < iter; i++ {
		entities = append(entities, ep.New())
	}

	for i := uint(0); i < iter; i++ {
		result := ep.Recycle(entities[i])
		assert.True(t, result, "expected Recycle() to return true for valid entities")
	}

	for i := iter - 1; i < iter; i-- {
		e := ep.New()
		assert.Equal(t, e.ID(), entities[i].ID(), "expected New() to return recycled items, (want %x, got %x)", entities[i], e)
		assert.Equal(t, e.Gen(), entities[i].Gen()+1, "expected New() to return new gen items, (want %x, got %x)", entities[i], e)
	}

}

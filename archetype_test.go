package ecs

import (
	"sort"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func testCheckArchetype(t *testing.T, arch *Archetype, components []EntityID) {
	assert.NotNil(t, arch, "Archetype should be valid")

	sort.Sort(EntityIDSlice(components))

	for index, id := range components {
		assert.Equal(t, arch.components[index], id, "Archetype with wrong component list (want %x, got %d)", id, arch.components[index])
		if !id.IsSingleton() {
			assert.NotNil(t, arch.columns[index], "components should have a Storage allocated")
		}
	}
}

func TestArchetypeGraph(t *testing.T) {
	type Position3D struct{ x, y, z float32 }
	type Orientation3D struct{ x, y, z float32 }
	type Health struct{ value float32 }
	type NameTag struct{ tag string }
	type Controlled struct{}

	factory := NewComponentFactory()
	entityPool := NewEntityPool(0)

	// This order matters for some tests that directly index the component Storage, don't change!
	posID := factory.Register(NewComponentRegistry[Position3D](entityPool))
	oriID := factory.Register(NewComponentRegistry[Orientation3D](entityPool))
	healthID := factory.Register(NewComponentRegistry[Health](entityPool))
	tagID := factory.Register(NewComponentRegistry[NameTag](entityPool))
	contrID := factory.Register(NewComponentSingletonRegistry[Controlled](entityPool))

	t.Run("ArchetypeGraphCore", func(t *testing.T) {
		ag := NewArchetypeGraph(NewComponentFactory())
		assert.Panics(t, func() {
			ag.Add(1000, 0, 1, 2, 3, 4, 5, 6)
		}, "trying to use invalid components should panic")

		assert.Panics(t, func() {
			ag.Add(1000)
			ag.Add(1000)
		}, "trying to add entity multiple times should panic")

		ag = NewArchetypeGraph(factory)

		assert.NotPanics(t, func() {
			ag.Rem(100)
			ag.Rem(0)
			ag.RemComponent(5, posID)
			ag.AddComponent(3, oriID)
			ag.Get(2)
		}, "invalid entities should not crash the archetype graph")

		e1 := entityPool.New()
		e2 := entityPool.New()
		e3 := entityPool.New()
		emptyEnt := entityPool.New()

		ag.Add(e1, healthID, posID, oriID)
		ag.Add(e2, posID, oriID)
		ag.Add(e3, posID, contrID, tagID)
		ag.Add(emptyEnt)

		rootArch, _ := ag.Get(emptyEnt)
		assert.NotNil(t, rootArch, "entities without components should be in the root archetype")

		posOriHealth, _ := ag.Get(e1)
		testCheckArchetype(t, posOriHealth, []EntityID{posID, oriID, healthID})

		posOri, _ := ag.Get(e2)
		testCheckArchetype(t, posOri, []EntityID{posID, oriID})

		posTagCtl, _ := ag.Get(e3)
		testCheckArchetype(t, posTagCtl, []EntityID{posID, tagID, contrID})

		ag.RemComponent(e1, healthID)
		archE1, _ := ag.Get(e1)
		assert.True(t, archE1 == posOri, "invalid archetype transition (want %d, got %d)", posOri.id, archE1.id)

		ag.AddComponent(e1, healthID)
		archE1, _ = ag.Get(e1)
		assert.True(t, archE1 == posOriHealth, "invalid archetype transition (want %d, got %d)", posOriHealth.id, archE1.id)

		ag.AddComponent(e3, posID)
		archE3, _ := ag.Get(e3)
		assert.True(t, posTagCtl == archE3, "adding component twice to entity should do nothing")

		ag.RemComponent(e3, oriID)
		archE3, _ = ag.Get(e3)
		assert.True(t, posTagCtl == archE3, "removing invalid components from entity should do nothing")

		ag.AddComponent(emptyEnt, tagID)
		tagArch, _ := ag.Get(emptyEnt)
		ag.Rem(e3)
		archE3, _ = ag.Get(e3)
		assert.Nil(t, archE3, "removed entities should not have archetype")

		ag.Add(e3, tagID)
		tagArch2, _ := ag.Get(e3)
		assert.True(t, tagArch == tagArch2, "archetypes should not be duplicated")

		ag.AddComponent(emptyEnt, oriID)
		oriTagArch, _ := ag.Get(emptyEnt)
		ag.AddComponent(e3, oriID)
		oriTagArch2, _ := ag.Get(e3)
		assert.True(t, oriTagArch == oriTagArch2, "entities with same components should be in the same archetype")
		assert.True(t, oriTagArch.entityCount == 2, "number of entities should be correct")
	})

	t.Run("ArchetypeGraph component values", func(t *testing.T) {
		ag := NewArchetypeGraph(factory)

		type EntityPos3D struct {
			EntityID
			Position3D
		}
		
		count := 10000
		
		pairs := make([]EntityPos3D, 0)
		for i := 0; i < count; i++ {
			pairs = append(pairs, EntityPos3D{entityPool.New(), Position3D{float32(i), float32(i), float32(i)}})
			comp := []EntityID{posID}
			if i % 3 == 1 {
				comp = append(comp, oriID)
			}
			if i % 3 == 2 {
				comp = append(comp, tagID)
			}
			ag.Add(pairs[i].EntityID, comp...)
		}

		for _, pair := range pairs {
			arch, row := ag.Get(pair.EntityID)
			arch.columns[0].Copy(uint(row), unsafe.Pointer(&pair.Position3D))
		}

		for _, pair := range pairs {
			arch, row := ag.Get(pair.EntityID)
			pos := (*Position3D)(arch.columns[0].Get(uint(row)))
			assert.Equal(t, *pos, pair.Position3D, "values mismatch (want %+v, got %+v)", pair.Position3D, *pos)
		}

		for i, pair := range pairs {
			if i % 3 == 2 {
				ag.RemComponent(pair.EntityID, tagID)
			}
			if i % 3 == 1 {
				ag.RemComponent(pair.EntityID, oriID)
				ag.AddComponent(pair.EntityID, tagID)
			}
			arch, row := ag.Get(pair.EntityID)
			pos := (*Position3D)(arch.columns[0].Get(uint(row)))
			assert.Equal(t, *pos, pair.Position3D, "values mismatch (want %+v, got %+v)", pair.Position3D, *pos)
		}

		for _, pair := range pairs {
			arch, row := ag.Get(pair.EntityID)
			pos := (*Position3D)(arch.columns[0].Get(uint(row)))
			assert.Equal(t, *pos, pair.Position3D, "values mismatch (want %+v, got %+v)", pair.Position3D, *pos)
		}
	})
}

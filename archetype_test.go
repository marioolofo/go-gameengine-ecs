//lint:file-ignore U1000 we are testing differente sizes of struct only
package ecs

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func testCheckArchetype(t *testing.T, arch *Archetype, components []ComponentID) {
	assert.NotNil(t, arch, "Archetype should be valid")

	for _, comp := range components {
		if arch.mask.IsSet(uint64(comp)) {
			assert.NotNil(t, arch.columns[comp], "Archetype with wrong component list (want %x, got %d)", MakeComponentMask(components...), arch.mask)
		}
	}
}

func TestArchetypeGraph(t *testing.T) {
	const (
		Pos3DCompID ComponentID = iota
		Ori3DCompID
		HealthCompID
		NameTagCompID
		ControlledCompID
	)
	type Position3D struct{ x, y, z float32 }
	type Orientation3D struct{ x, y, z float32 }
	type Health struct{ value float32 }
	type NameTag struct{ tag string }
	type Controlled struct{}

	factory := NewComponentFactory()
	entityPool := NewEntityPool(0)

	// This order matters for some tests that directly index the component Storage, don't change!
	factory.Register(NewComponentRegistry[Position3D](Pos3DCompID))
	factory.Register(NewComponentRegistry[Orientation3D](Ori3DCompID))
	factory.Register(NewComponentRegistry[Health](HealthCompID))
	factory.Register(NewComponentRegistry[NameTag](NameTagCompID))
	factory.Register(NewComponentRegistry[Controlled](ControlledCompID))

	t.Run("ArchetypeGraphCore", func(t *testing.T) {
		compFac := NewComponentFactory()
		ag := NewArchetypeGraph(compFac)
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
			ag.RemComponent(5, Pos3DCompID)
			ag.AddComponent(3, Ori3DCompID)
			ag.Get(2)
		}, "invalid entities should not crash the archetype graph")

		e1 := entityPool.New()
		e2 := entityPool.New()
		e3 := entityPool.New()
		emptyEnt := entityPool.New()

		ag.Add(e1, HealthCompID, Pos3DCompID, Ori3DCompID)
		ag.Add(e2, Pos3DCompID, Ori3DCompID)
		ag.Add(e3, Pos3DCompID, ControlledCompID, NameTagCompID)
		ag.Add(emptyEnt)

		rootArch, _ := ag.Get(emptyEnt)
		assert.NotNil(t, rootArch, "entities without components should be in the root archetype")

		posOriHealth, _ := ag.Get(e1)
		testCheckArchetype(t, posOriHealth, []ComponentID{Pos3DCompID, Ori3DCompID, HealthCompID})

		posOri, _ := ag.Get(e2)
		testCheckArchetype(t, posOri, []ComponentID{Pos3DCompID, Ori3DCompID})

		posTagCtl, _ := ag.Get(e3)
		testCheckArchetype(t, posTagCtl, []ComponentID{Pos3DCompID, NameTagCompID, ControlledCompID})

		ag.RemComponent(e1, HealthCompID)
		archE1, _ := ag.Get(e1)
		assert.True(t, archE1 == posOri, "invalid archetype transition (want %d, got %d)", posOri.mask, archE1.mask)

		ag.AddComponent(e1, HealthCompID)
		archE1, _ = ag.Get(e1)
		assert.True(t, archE1 == posOriHealth, "invalid archetype transition (want %d, got %d)", posOriHealth.mask, archE1.mask)
		assert.True(t, archE1.Component(HealthCompID, 0) != unsafe.Pointer(nil), "archetype.Component should return valid pointer")

		ag.AddComponent(e3, Pos3DCompID)
		archE3, _ := ag.Get(e3)
		assert.True(t, posTagCtl == archE3, "adding component twice to entity should do nothing")

		ag.RemComponent(e3, Ori3DCompID)
		archE3, _ = ag.Get(e3)
		assert.True(t, posTagCtl == archE3, "removing invalid components from entity should do nothing")

		ag.AddComponent(emptyEnt, NameTagCompID)
		tagArch, _ := ag.Get(emptyEnt)
		ag.Rem(e3)
		archE3, _ = ag.Get(e3)
		assert.Nil(t, archE3, "removed entities should not have archetype")

		ag.Add(e3, NameTagCompID)
		tagArch2, _ := ag.Get(e3)
		assert.True(t, tagArch == tagArch2, "archetypes should not be duplicated")

		ag.AddComponent(emptyEnt, Ori3DCompID)
		oriTagArch, _ := ag.Get(emptyEnt)
		ag.AddComponent(e3, Ori3DCompID)
		oriTagArch2, _ := ag.Get(e3)
		assert.True(t, oriTagArch == oriTagArch2, "entities with same components should be in the same archetype")
		assert.True(t, len(oriTagArch.entities) == 2, "number of entities should be correct")
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
			comp := []ComponentID{Pos3DCompID}
			if i%3 == 1 {
				comp = append(comp, Ori3DCompID)
			}
			if i%3 == 2 {
				comp = append(comp, NameTagCompID)
			}
			ag.Add(pairs[i].EntityID, comp...)
		}

		for _, pair := range pairs {
			arch, row := ag.Get(pair.EntityID)
			arch.columns[Pos3DCompID].Copy(uint(row), unsafe.Pointer(&pair.Position3D))
		}

		for _, pair := range pairs {
			arch, row := ag.Get(pair.EntityID)
			pos := (*Position3D)(arch.columns[Pos3DCompID].Get(uint(row)))
			assert.Equal(t, *pos, pair.Position3D, "values mismatch (want %+v, got %+v)", pair.Position3D, *pos)
		}

		for i, pair := range pairs {
			if i%3 == 2 {
				ag.RemComponent(pair.EntityID, NameTagCompID)
			}
			if i%3 == 1 {
				ag.RemComponent(pair.EntityID, Ori3DCompID)
				ag.AddComponent(pair.EntityID, NameTagCompID)
			}
			arch, row := ag.Get(pair.EntityID)
			pos := (*Position3D)(arch.columns[Pos3DCompID].Get(uint(row)))
			assert.Equal(t, *pos, pair.Position3D, "values mismatch (want %+v, got %+v)", pair.Position3D, *pos)
		}

		for _, pair := range pairs {
			arch, row := ag.Get(pair.EntityID)
			pos := (*Position3D)(arch.columns[Pos3DCompID].Get(uint(row)))
			assert.Equal(t, *pos, pair.Position3D, "values mismatch (want %+v, got %+v)", pair.Position3D, *pos)
		}
	})
}

func BenchmarkArchetypeGraph(b *testing.B) {
	entityPool := NewEntityPool(100000)
	factory := NewComponentFactory()
	graph := NewArchetypeGraph(factory)

	type Vec2D struct {
		x, y float32
	}
	type UIDesign struct {
		name  string
		flags uint64
	}
	type Transform2D struct {
		position Vec2D
		rotation float32
	}
	type Physics2D struct {
		linearAccel, velocity Vec2D
		angularAccel, torque  float32
	}
	type Script struct {
		handle int
	}

	const (
		Vec2DCompID ComponentID = iota
		UIDesignCompID
		Transf2DCompID
		Phys2DCompID
		ScriptCompID
	)

	factory.Register(NewComponentRegistry[UIDesign](UIDesignCompID))
	factory.Register(NewComponentRegistry[Transform2D](Transf2DCompID))
	factory.Register(NewComponentRegistry[Physics2D](Phys2DCompID))
	factory.Register(NewComponentRegistry[Script](ScriptCompID))

	entityCount := 10000
	updateCount := b.N

	for i := 0; i < entityCount/2; i++ {
		e1 := entityPool.New()
		graph.Add(e1, UIDesignCompID, ScriptCompID)

		arch, row := graph.Get(e1)

		design := (*UIDesign)(arch.Component(UIDesignCompID, row))
		design.name = fmt.Sprint("entity_", i)

		e2 := entityPool.New()
		graph.Add(e2, Transf2DCompID, Phys2DCompID)

		trArch, row := graph.Get(e2)
		phys := (*Physics2D)(trArch.Component(Phys2DCompID, row))
		phys.linearAccel = Vec2D{x: 2, y: 1.5}
	}

	mask := MakeComponentMask(Transf2DCompID, Phys2DCompID)

	dt := float32(1.0 / 60.0)

	count := 0

	for i := 0; i < updateCount; i++ {
		iter := graph.Query(mask)
		for iter.Next() {
			count++
			tr := (*Transform2D)(iter.Component(Transf2DCompID))
			phys := (*Physics2D)(iter.Component(Phys2DCompID))

			phys.velocity.x += phys.linearAccel.x * dt
			phys.velocity.y += phys.linearAccel.y * dt

			tr.position.x += phys.velocity.x * dt
			tr.position.y += phys.velocity.y * dt

			phys.velocity.x *= 0.99
			phys.velocity.y *= 0.99
		}
	}
	b.Logf("[geecs] found %d items to query", count)
}

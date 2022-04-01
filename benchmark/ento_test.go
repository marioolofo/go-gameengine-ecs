package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/wfranczyk/ento"
)

type SystemPhysTransform struct {
	Phys *Physics2D `ento:"required"`
	Transform *Transform2D `ento:"required"`
}

type SystemDesignScript struct {
	Design *UIDesign `ento:"required"`
	Script *Script `ento:"required"`
}

func (d *SystemDesignScript) Update(entity *ento.Entity) {
}

func (s *SystemPhysTransform) Update(entity *ento.Entity) {

	var phys *Physics2D
	var tr *Transform2D

	entity.Get(&phys)
	entity.Get(&tr)

	phys.velocity.x += phys.linearAccel.x * dt
	phys.velocity.y += phys.linearAccel.y * dt

	tr.position.x += phys.velocity.x * dt
	tr.position.y += phys.velocity.y * dt

	phys.velocity.x *= 0.99
	phys.velocity.y *= 0.99
}

func EntoBench(b *testing.B, entityCount, updateCount int) {
	world := ento.NewWorldBuilder().
		WithSparseComponents(Physics2D{}, Transform2D{}, Script{}, UIDesign{}).Build(256)

	physSystem := &SystemPhysTransform{}
	uiSystem := &SystemDesignScript{}

	world.AddSystems(physSystem, uiSystem)

	for i := 0; i < entityCount / 2; i++ {
		name := fmt.Sprint("entity_", i)
		world.AddEntity(UIDesign{name: name}, Script{})

		world.AddEntity(Transform2D{}, Physics2D{linearAccel: Vec2D{x: 2, y: 1.5}})
	}

	for i := 0; i < updateCount; i++ {
		world.Update()
	}
}

// 0 updates

func BenchmarkEnto_100_entities_0_updates(b *testing.B) {
	EntoBench(b, 100, 0)
}

func BenchmarkEnto_1000_entities_0_updates(b *testing.B) {
	EntoBench(b, 1000, 0)
}

func BenchmarkEnto_10000_entities_0_updates(b *testing.B) {
	EntoBench(b, 10000, 0)
}

func BenchmarkEnto_100000_entities_0_updates(b *testing.B) {
	EntoBench(b, 100000, 0)
}

// 100 updates

func BenchmarkEnto_100_entities_100_updates(b *testing.B) {
	EntoBench(b, 100, 100)
}

func BenchmarkEnto_1000_entities_100_updates(b *testing.B) {
	EntoBench(b, 1000, 100)
}

func BenchmarkEnto_10000_entities_100_updates(b *testing.B) {
	EntoBench(b, 10000, 100)
}

func BenchmarkEnto_100000_entities_100_updates(b *testing.B) {
	EntoBench(b, 100000, 100)
}

// 1000 updates

func BenchmarkEnto_100_entities_1000_updates(b *testing.B) {
	EntoBench(b, 100, 1000)
}

func BenchmarkEnto_1000_entities_1000_updates(b *testing.B) {
	EntoBench(b, 1000, 1000)
}

func BenchmarkEnto_10000_entities_1000_updates(b *testing.B) {
	EntoBench(b, 10000, 1000)
}

func BenchmarkEnto_100000_entities_1000_updates(b *testing.B) {
	EntoBench(b, 100000, 1000)
}

// 10000 updates

func BenchmarkEnto_100_entities_10000_updates(b *testing.B) {
	EntoBench(b, 100, 10000)
}

func BenchmarkEnto_1000_entities_10000_updates(b *testing.B) {
	EntoBench(b, 1000, 10000)
}

func BenchmarkEnto_10000_entities_10000_updates(b *testing.B) {
	EntoBench(b, 10000, 10000)
}

func BenchmarkEnto_100000_entities_10000_updates(b *testing.B) {
	EntoBench(b, 100000, 10000)
}


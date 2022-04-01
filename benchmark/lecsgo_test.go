package ecs_benchmark

import (
	"fmt"
	"testing"
	_ "github.com/leopotam/go-ecs"
)

func LecsGOBench(b *testing.B, entityCount, updateCount int) {
	world := NewLecsGOWorld(256)

	for i := 0; i < entityCount / 2; i++ {
		e1 := world.NewEntity()
		world.SetScript(e1)
		design := world.SetUIDesign(e1)
		design.name = fmt.Sprint("entity_", i)

		e2 := world.NewEntity()
		world.SetTransform2D(e2)
		phys := world.SetPhysics2D(e2)
		phys.velocity = Vec2D{x: 2, y: 1.5}
	}

	for i := 0; i < updateCount; i++ {
		for _, entity := range world.WithPhysTransform().Entities() {
			phys := world.GetPhysics2D(entity)
			tr := world.GetTransform2D(entity)

			phys.velocity.x += phys.linearAccel.x * dt
			phys.velocity.y += phys.linearAccel.y * dt

			tr.position.x += phys.velocity.x * dt
			tr.position.y += phys.velocity.y * dt

			phys.velocity.x *= 0.99
			phys.velocity.y *= 0.99
		}
	}
	world.Destroy()
}

// 0 updates

func BenchmarkLecsGO_100_entities_0_updates(b *testing.B) {
	LecsGOBench(b, 100, 0)
}

func BenchmarkLecsGO_1000_entities_0_updates(b *testing.B) {
	LecsGOBench(b, 1000, 0)
}

func BenchmarkLecsGO_10000_entities_0_updates(b *testing.B) {
	LecsGOBench(b, 10000, 0)
}

func BenchmarkLecsGO_100000_entities_0_updates(b *testing.B) {
	LecsGOBench(b, 100000, 0)
}

// 100 updates

func BenchmarkLecsGO_100_entities_100_updates(b *testing.B) {
	LecsGOBench(b, 100, 100)
}

func BenchmarkLecsGO_1000_entities_100_updates(b *testing.B) {
	LecsGOBench(b, 1000, 100)
}

func BenchmarkLecsGO_10000_entities_100_updates(b *testing.B) {
	LecsGOBench(b, 10000, 100)
}

func BenchmarkLecsGO_100000_entities_100_updates(b *testing.B) {
	LecsGOBench(b, 100000, 100)
}

// 1000 updates

func BenchmarkLecsGO_100_entities_1000_updates(b *testing.B) {
	LecsGOBench(b, 100, 1000)
}

func BenchmarkLecsGO_1000_entities_1000_updates(b *testing.B) {
	LecsGOBench(b, 1000, 1000)
}

func BenchmarkLecsGO_10000_entities_1000_updates(b *testing.B) {
	LecsGOBench(b, 10000, 1000)
}

func BenchmarkLecsGO_100000_entities_1000_updates(b *testing.B) {
	LecsGOBench(b, 100000, 1000)
}

// 10000 updates

func BenchmarkLecsGO_100_entities_10000_updates(b *testing.B) {
	LecsGOBench(b, 100, 10000)
}

func BenchmarkLecsGO_1000_entities_10000_updates(b *testing.B) {
	LecsGOBench(b, 1000, 10000)
}

func BenchmarkLecsGO_10000_entities_10000_updates(b *testing.B) {
	LecsGOBench(b, 10000, 10000)
}

func BenchmarkLecsGO_100000_entities_10000_updates(b *testing.B) {
	LecsGOBench(b, 100000, 10000)
}


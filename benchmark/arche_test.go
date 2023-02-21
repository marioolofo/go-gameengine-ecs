package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/mlange-42/arche/ecs"
)

func ArcheECSBench(b *testing.B, entityCount, updateCount int) {
	// Create a World.
	world := ecs.NewWorld()

	uidesignComponentID := ecs.ComponentID[UIDesign](&world)
	transformComponentID := ecs.ComponentID[Transform2D](&world)
	physicsComponentID := ecs.ComponentID[Physics2D](&world)
	scriptComponentID := ecs.ComponentID[Script](&world)

	uidesignScriptComponents := []ecs.ID{uidesignComponentID, scriptComponentID}
	transfPhysComponents := []ecs.ID{transformComponentID, physicsComponentID}

	// Create entities.
	for i := 0; i < entityCount/2; i++ {
		// Create an Entity wth these components.
		e1 := world.NewEntity()
		world.Add(e1, uidesignScriptComponents...)
		design := (*UIDesign)(world.Get(e1, uidesignComponentID))
		design.name = fmt.Sprint("entity_", i)

		e2 := world.NewEntity()
		world.Add(e2, transfPhysComponents...)

		phys := (*Physics2D)(world.Get(e2, physicsComponentID))
		phys.linearAccel = Vec2D{x: 2, y: 1.5}
	}

	for i := 0; i < updateCount; i++ {
		// Get a fresh query iterator.
		query := world.Query(transformComponentID, physicsComponentID)
		// Iterate it.
		for query.Next() {
			// Component access through the Query.
			tr := (*Transform2D)(query.Get(transformComponentID))
			phys := (*Physics2D)(query.Get(physicsComponentID))
			// Update component fields.
			phys.velocity.x += phys.linearAccel.x * dt
			phys.velocity.y += phys.linearAccel.y * dt

			tr.position.x += phys.velocity.x * dt
			tr.position.y += phys.velocity.y * dt

			phys.velocity.x *= 0.99
			phys.velocity.y *= 0.99
		}
	}
}

// 0 updates

func BenchmarkArcheECS_100_entities_0_updates(b *testing.B) {
	ArcheECSBench(b, 100, 0)
}

func BenchmarkArcheECS_1000_entities_0_updates(b *testing.B) {
	ArcheECSBench(b, 1000, 0)
}

func BenchmarkArcheECS_10000_entities_0_updates(b *testing.B) {
	ArcheECSBench(b, 10000, 0)
}

func BenchmarkArcheECS_100000_entities_0_updates(b *testing.B) {
	ArcheECSBench(b, 100000, 0)
}

// 100 updates

func BenchmarkArcheECS_100_entities_100_updates(b *testing.B) {
	ArcheECSBench(b, 100, 100)
}

func BenchmarkArcheECS_1000_entities_100_updates(b *testing.B) {
	ArcheECSBench(b, 1000, 100)
}

func BenchmarkArcheECS_10000_entities_100_updates(b *testing.B) {
	ArcheECSBench(b, 10000, 100)
}

func BenchmarkArcheECS_100000_entities_100_updates(b *testing.B) {
	ArcheECSBench(b, 100000, 100)
}

// 1000 updates

func BenchmarkArcheECS_100_entities_1000_updates(b *testing.B) {
	ArcheECSBench(b, 100, 1000)
}

func BenchmarkArcheECS_1000_entities_1000_updates(b *testing.B) {
	ArcheECSBench(b, 1000, 1000)
}

func BenchmarkArcheECS_10000_entities_1000_updates(b *testing.B) {
	ArcheECSBench(b, 10000, 1000)
}

func BenchmarkArcheECS_100000_entities_1000_updates(b *testing.B) {
	ArcheECSBench(b, 100000, 1000)
}

// 10000 updates

func BenchmarkArcheECS_100_entities_10000_updates(b *testing.B) {
	ArcheECSBench(b, 100, 10000)
}

func BenchmarkArcheECS_1000_entities_10000_updates(b *testing.B) {
	ArcheECSBench(b, 1000, 10000)
}

func BenchmarkArcheECS_10000_entities_10000_updates(b *testing.B) {
	ArcheECSBench(b, 10000, 10000)
}

func BenchmarkArcheECS_100000_entities_10000_updates(b *testing.B) {
	ArcheECSBench(b, 100000, 10000)
}

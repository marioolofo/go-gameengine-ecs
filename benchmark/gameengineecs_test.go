package ecs_benchmark

import (
	"fmt"
	"testing"

	ecs "github.com/marioolofo/go-gameengine-ecs"
)

func GameEngineECSBench(b *testing.B, entityCount, updateCount int) {
	entityPool := ecs.NewEntityPool(100000)
	factory := ecs.NewComponentFactory()
	graph := ecs.NewArchetypeGraph(factory)

	uidesignComponentID := factory.Register(ecs.NewComponentRegistry[UIDesign](entityPool))
	transformComponentID := factory.Register(ecs.NewComponentRegistry[Transform2D](entityPool))
	physicsComponentID := factory.Register(ecs.NewComponentRegistry[Physics2D](entityPool))
	scriptComponentID := factory.Register(ecs.NewComponentRegistry[Script](entityPool))

	entities := make([]ecs.EntityID, entityCount/2)

	for i := 0; i < entityCount/2; i++ {
		e1 := entityPool.New()
		graph.Add(e1, uidesignComponentID, scriptComponentID)

		arch, row := graph.Get(e1)

		design := (*UIDesign)(arch.GetComponentPtr(0, row))
		design.name = fmt.Sprint("entity_", i)

		e2 := entityPool.New()
		graph.Add(e2, transformComponentID, physicsComponentID)

		entities[i] = e2

		trArch, row := graph.Get(e2)
		phys := (*Physics2D)(trArch.GetComponentPtr(1, row))
		phys.linearAccel = Vec2D{x: 2, y: 1.5}
	}

	for i := 0; i < updateCount; i++ {
		for _, entity := range entities {
			trArch, row := graph.Get(entity)
			tr := (*Transform2D)(trArch.GetComponentPtr(0, row))
			phys := (*Physics2D)(trArch.GetComponentPtr(1, row))

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

func BenchmarkGameEngineECS_100_entities_0_updates(b *testing.B) {
	GameEngineECSBench(b, 100, 0)
}

func BenchmarkGameEngineECS_1000_entities_0_updates(b *testing.B) {
	GameEngineECSBench(b, 1000, 0)
}

func BenchmarkGameEngineECS_10000_entities_0_updates(b *testing.B) {
	GameEngineECSBench(b, 10000, 0)
}

func BenchmarkGameEngineECS_100000_entities_0_updates(b *testing.B) {
	GameEngineECSBench(b, 100000, 0)
}

// 100 updates

func BenchmarkGameEngineECS_100_entities_100_updates(b *testing.B) {
	GameEngineECSBench(b, 100, 100)
}

func BenchmarkGameEngineECS_1000_entities_100_updates(b *testing.B) {
	GameEngineECSBench(b, 1000, 100)
}

func BenchmarkGameEngineECS_10000_entities_100_updates(b *testing.B) {
	GameEngineECSBench(b, 10000, 100)
}

func BenchmarkGameEngineECS_100000_entities_100_updates(b *testing.B) {
	GameEngineECSBench(b, 100000, 100)
}

// 1000 updates

func BenchmarkGameEngineECS_100_entities_1000_updates(b *testing.B) {
	GameEngineECSBench(b, 100, 1000)
}

func BenchmarkGameEngineECS_1000_entities_1000_updates(b *testing.B) {
	GameEngineECSBench(b, 1000, 1000)
}

func BenchmarkGameEngineECS_10000_entities_1000_updates(b *testing.B) {
	GameEngineECSBench(b, 10000, 1000)
}

func BenchmarkGameEngineECS_100000_entities_1000_updates(b *testing.B) {
	GameEngineECSBench(b, 100000, 1000)
}

// 10000 updates

func BenchmarkGameEngineECS_100_entities_10000_updates(b *testing.B) {
	GameEngineECSBench(b, 100, 10000)
}

func BenchmarkGameEngineECS_1000_entities_10000_updates(b *testing.B) {
	GameEngineECSBench(b, 1000, 10000)
}

func BenchmarkGameEngineECS_10000_entities_10000_updates(b *testing.B) {
	GameEngineECSBench(b, 10000, 10000)
}

func BenchmarkGameEngineECS_100000_entities_10000_updates(b *testing.B) {
	GameEngineECSBench(b, 100000, 10000)
}

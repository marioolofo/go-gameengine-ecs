package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/marioolofo/go-gameengine-ecs"
)

func GameEngineECSBench(b *testing.B, entityCount, updateCount int) {
	config := []ecs.ComponentConfig{
		{ID: UIDesignComponentID, Component: UIDesign{}},
		{ID: Transform2DComponentID, Component: Transform2D{}},
		{ID: Physics2DComponentID, Component: Physics2D{}},
		{ID: ScriptComponentID, Component: Script{}},
	}

	world := ecs.NewWorld(config...)

	for i := 0; i < entityCount/2; i++ {
		e1 := world.NewEntity()
		world.Assign(e1, UIDesignComponentID, ScriptComponentID)
		design := (*UIDesign)(world.Component(e1, UIDesignComponentID))
		design.name = fmt.Sprint("entity_", i)

		e2 := world.NewEntity()
		world.Assign(e2, Transform2DComponentID, Physics2DComponentID)
		phys := (*Physics2D)(world.Component(e2, Physics2DComponentID))
		phys.linearAccel = Vec2D{x: 2, y: 1.5}
	}

	var filter ecs.Filter
	if BenchUpdateCount > 0 {
		filter = world.NewFilter(Transform2DComponentID, Physics2DComponentID)
	}

	for i := 0; i < updateCount; i++ {
		for _, entity := range filter.Entities() {
			phys := (*Physics2D)(world.Component(entity, Physics2DComponentID))
			tr := (*Transform2D)(world.Component(entity, Transform2DComponentID))

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

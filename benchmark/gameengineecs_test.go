package ecs_benchmark

import (
	"fmt"
	"testing"

	ecs "github.com/marioolofo/go-gameengine-ecs"
)

func GameEngineECSBench(b *testing.B, entityCount, updateCount int) {
	world := ecs.NewWorld(uint(entityCount + 10))

	world.Register(ecs.NewComponentRegistry[UIDesign](UIDesignComponentID))
	world.Register(ecs.NewComponentRegistry[Transform2D](Transform2DComponentID))
	world.Register(ecs.NewComponentRegistry[Physics2D](Physics2DComponentID))
	world.Register(ecs.NewComponentRegistry[Script](ScriptComponentID))

	for i := 0; i < entityCount/2; i++ {
		e1 := world.NewEntity(UIDesignComponentID, ScriptComponentID)

		design := (*UIDesign)(world.GetComponent(e1, UIDesignComponentID))
		design.name = fmt.Sprint("entity_", i)

		e2 := world.NewEntity(Transform2DComponentID, Physics2DComponentID)

		phys := (*Physics2D)(world.GetComponent(e2, Physics2DComponentID))
		phys.linearAccel = Vec2D{x: 2, y: 1.5}
	}

	mask := ecs.MakeComponentMask(Transform2DComponentID, Physics2DComponentID)

	dt := float32(1.0 / 60.0)

	for i := 0; i < updateCount; i++ {
		iter := world.Query(mask)
		for iter.Next() {
			tr := (*Transform2D)(iter.Get(Transform2DComponentID))
			phys := (*Physics2D)(iter.Get(Physics2DComponentID))

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

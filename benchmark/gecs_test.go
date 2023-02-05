package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/tutumagi/gecs"
)

var (
	PhysID      = gecs.RegisterComponent("Phys2D")
	TransformID = gecs.RegisterComponent("Transform2D")
	ScriptID    = gecs.RegisterComponent("Script")
	UIDesignID  = gecs.RegisterComponent("UIDesign")
)

func GecsBench(b *testing.B, entityCount, updateCount int) {
	registry := gecs.NewRegistry()

	for i := 0; i < entityCount/2; i++ {
		name := fmt.Sprint("entity_", i)
		e1 := registry.Create()
		registry.Assign(e1, ScriptID, Script{})
		registry.Assign(e1, UIDesignID, UIDesign{name: name})

		e2 := registry.Create()
		registry.Assign(e2, TransformID, Transform2D{})
		registry.Assign(e2, PhysID, Physics2D{linearAccel: Vec2D{x: 2, y: 1.5}})
	}

	for i := 0; i < updateCount; i++ {
		registry.View(TransformID, PhysID).Each(func(e gecs.EntityID, data map[gecs.ComponentID]interface{}) {
			phys := data[PhysID].(Physics2D)
			tr := data[TransformID].(Transform2D)

			phys.velocity.x += phys.linearAccel.x * dt
			phys.velocity.y += phys.linearAccel.y * dt

			tr.position.x += phys.velocity.x * dt
			tr.position.y += phys.velocity.y * dt

			phys.velocity.x *= 0.99
			phys.velocity.y *= 0.99
		})
	}
}

// 0 updates

func BenchmarkGecs_100_entities_0_updates(b *testing.B) {
	GecsBench(b, 100, 0)
}

func BenchmarkGecs_1000_entities_0_updates(b *testing.B) {
	GecsBench(b, 1000, 0)
}

func BenchmarkGecs_10000_entities_0_updates(b *testing.B) {
	GecsBench(b, 10000, 0)
}

func BenchmarkGecs_100000_entities_0_updates(b *testing.B) {
	GecsBench(b, 100000, 0)
}

// 100 updates

func BenchmarkGecs_100_entities_100_updates(b *testing.B) {
	GecsBench(b, 100, 100)
}

func BenchmarkGecs_1000_entities_100_updates(b *testing.B) {
	GecsBench(b, 1000, 100)
}

func BenchmarkGecs_10000_entities_100_updates(b *testing.B) {
	GecsBench(b, 10000, 100)
}

func BenchmarkGecs_100000_entities_100_updates(b *testing.B) {
	GecsBench(b, 100000, 100)
}

// 1000 updates

func BenchmarkGecs_100_entities_1000_updates(b *testing.B) {
	GecsBench(b, 100, 1000)
}

func BenchmarkGecs_1000_entities_1000_updates(b *testing.B) {
	GecsBench(b, 1000, 1000)
}

func BenchmarkGecs_10000_entities_1000_updates(b *testing.B) {
	GecsBench(b, 10000, 1000)
}

func BenchmarkGecs_100000_entities_1000_updates(b *testing.B) {
	GecsBench(b, 100000, 1000)
}

// 10000 updates

func BenchmarkGecs_100_entities_10000_updates(b *testing.B) {
	GecsBench(b, 100, 10000)
}

func BenchmarkGecs_1000_entities_10000_updates(b *testing.B) {
	GecsBench(b, 1000, 10000)
}

func BenchmarkGecs_10000_entities_10000_updates(b *testing.B) {
	GecsBench(b, 10000, 10000)
}

func BenchmarkGecs_100000_entities_10000_updates(b *testing.B) {
	GecsBench(b, 100000, 10000)
}

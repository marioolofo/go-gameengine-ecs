package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/tutumagi/gecs"
)

var (
	PhysID = gecs.RegisterComponent("Phys2D")
	TransformID = gecs.RegisterComponent("Transform2D")
	ScriptID = gecs.RegisterComponent("Script")
	UIDesignID = gecs.RegisterComponent("UIDesign")
)

func BenchmarkGecs(b *testing.B) {
	registry := gecs.NewRegistry()

	for i := 0; i < BenchEntityCount; i++ {
		name := fmt.Sprint("entity_", i)
		e1 := registry.Create()
		registry.Assign(e1, ScriptID, Script{})
		registry.Assign(e1, UIDesignID, UIDesign{name: name})

		e2 := registry.Create()
		registry.Assign(e2, TransformID, Transform2D{})
		registry.Assign(e2, PhysID, Physics2D{linearAccel: Vec2D{x: 2, y: 1.5}})
	}

	for i := 0; i < BenchUpdateCount; i++ {
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

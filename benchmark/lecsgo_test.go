package ecs_benchmark

import (
	"fmt"
	"testing"
	_ "github.com/leopotam/go-ecs"
)

func BenchmarkLecsGO(b *testing.B) {
	world := NewLecsGOWorld(256)

	for i := 0; i < BenchEntityCount; i++ {
		e1 := world.NewEntity()
		world.SetScript(e1)
		design := world.SetUIDesign(e1)
		design.name = fmt.Sprint("entity_", i)

		e2 := world.NewEntity()
		world.SetTransform2D(e2)
		phys := world.SetPhysics2D(e2)
		phys.velocity = Vec2D{x: 2, y: 1.5}
	}

	for i := 0; i < BenchUpdateCount; i++ {
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

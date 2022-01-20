package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/marioolofo/go-gameengine-ecs"
)

func BenchmarkGameEngineECS(b *testing.B) {
	config := []ecs.ComponentConfig{
		{ID: UIDesignComponentID, Component: UIDesign{}},
		{ID: Transform2DComponentID, Component: Transform2D{}},
		{ID: Physics2DComponentID, Component: Physics2D{}},
		{ID: ScriptComponentID, Component: Script{}},
	}

	world := ecs.NewWorld(config...)

	for i := 0; i < BenchEntityCount; i++ {
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

	for i := 0; i < BenchUpdateCount; i++ {
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

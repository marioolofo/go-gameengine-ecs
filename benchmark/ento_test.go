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

func BenchmarkEnto(b *testing.B) {
	world := ento.NewWorldBuilder().
		WithSparseComponents(Physics2D{}, Transform2D{}, Script{}, UIDesign{}).Build(256)

	physSystem := &SystemPhysTransform{}
	uiSystem := &SystemDesignScript{}

	world.AddSystems(physSystem, uiSystem)

	for i := 0; i < BenchEntityCount; i++ {
		name := fmt.Sprint("entity_", i)
		world.AddEntity(UIDesign{name: name}, Script{})

		world.AddEntity(Transform2D{}, Physics2D{linearAccel: Vec2D{x: 2, y: 1.5}})
	}

	for i := 0; i < BenchUpdateCount; i++ {
		world.Update()
	}

}

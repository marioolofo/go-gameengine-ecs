package ecs_benchmark

import (
	"fmt"
	"testing"

	ecs "github.com/marioolofo/go-gameengine-ecs/benchmark/Entitas"
)

type Transform struct {
	group ecs.Group
}

func (t *Transform) Initer(contexts ecs.Contexts) {
	game := contexts.Entitas()
	t.group = game.Group(ecs.NewMatcher().AllOf(ecs.EPhysics2D, ecs.ETransform2D))
}

func (t *Transform) Executer() {
	for _, e := range t.group.GetEntities() {
		phys := e.GetEPhysics2D()
		tr := e.GetETransform2D()

		Velx := phys.Velx + phys.Accelx * dt
		Vely := phys.Vely + phys.Accely * dt

		X := tr.X + Velx * dt
		Y := tr.Y + Vely * dt

		phys.Velx *= 0.99
		phys.Vely *= 0.99

		e.ReplaceEPhysics2D(phys.Accely, Velx, Vely, phys.AngularAccel, phys.Torque, phys.Accelx)
		e.ReplaceETransform2D(X, Y, tr.Rotation)
	}
}

func BenchmarkEntitas(b *testing.B) {
	contexts := ecs.CreateContexts()
	game := contexts.Entitas()

	systems := ecs.CreateSystemPool()
	systems.Add(&Transform{})

	for i := 0; i < BenchEntityCount; i++ {
		name := fmt.Sprint("entity_", i)
		e1 := game.CreateEntity()
		e1.AddEScript(0)
		e1.AddEUIDesign(name, 0)

		e2 := game.CreateEntity()
		e2.AddETransform2D(0, 0, 0)
		e2.AddEPhysics2D(0, 0, 2, 1.5, 0, 0)
	}

	systems.Init(contexts)

	for i := 0; i < BenchUpdateCount; i++ {
		systems.Execute()
		systems.Clean()
	}
}

package ecs_benchmark

import (
	ecs "github.com/leopotam/go-ecs"
)

type lecsgoWorld interface {
	components() (
		UIDesign,
		Script,
		Transform2D,
		Physics2D,
	)

	WithPhysTransform(tr Transform2D, phys Physics2D)
	WithDesignScript(ui UIDesign, script Script)
}

type LecsGOWorld struct {
	world *ecs.World `ecs:"lecsgoWorld"`
}

const lecsgoWorldName = "lecsgoWorld"

//go:generate go run github.com/leopotam/go-ecs/cmd/world-gen

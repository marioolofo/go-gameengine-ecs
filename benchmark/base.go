package ecs_benchmark

import (
	ecs "github.com/marioolofo/go-gameengine-ecs"
)

const (
	UIDesignComponentID ecs.EntityID = iota
	Transform2DComponentID
	Physics2DComponentID
	ScriptComponentID
	CustomComponentStartID
)

const (
	BenchEntityCount = 1000
	BenchUpdateCount = 1000

	dt = float32(1.0 / 60.0)
)

type Vec2D struct {
	x, y float32
}

type UIDesign struct {
	name  string
	flags uint64
}

type Transform2D struct {
	position Vec2D
	rotation float32
}

type Physics2D struct {
	linearAccel, velocity Vec2D
	angularAccel, torque  float32
}

type Script struct {
	handle int
}

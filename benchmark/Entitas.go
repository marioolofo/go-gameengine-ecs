package ecs_benchmark

//go:generate go run github.com/Falldot/Entitas-Go

type EUIDesign struct {
	Name   string
	Flags  uint64
}

type ETransform2D struct {
	X, Y float32
	Rotation float32
}

type EPhysics2D struct {
	Accelx, Accely, Velx, Vely float32
	AngularAccel, Torque float32
}

type EScript struct {
	Handle int
}


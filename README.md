# ECS

## _Fast Entity Component System in Golang_

This module is the ECS part of the game engine i'm writing in Go.

Features:

- as fast as packages with automatic code generation, but no setup and regeneration required for every change
- bitmask for component mapping
- sparse arrays for fast memory mapping
- memory allocated in chunks instead of dynamic arrays, ensuring a single memory
  address for the lifetime of the component instance
- filters instead of services, automatically updated after adding or removing
  components to the world
- the code is commented and the documentation can be generated with godoc
- 100% test coverage

### Installation

```sh
go get github.com/marioolofo/go-gameengine-ecs
```

> This project was made with Go 1.17, but it may work with older versions too

## Example

See the [examples](./examples) folder.

From [simple.go](./examples/simple/simple.ho):

```go
package main

import (
	"github.com/marioolofo/go/gameengine/ecs"
)

// Component IDs
const (
	TransformID ecs.ID = iota
	PhysicsID
)

type Vec2D struct {
	x, y float32
}

type TransformComponent struct {
	position Vec2D
	rotation float32
}

type PhysicsComponent struct {
	linearAccel, velocity Vec2D
	angularAccel, torque  float32
}

func main() {
	// initial configuration to create the world, new components can be
	// added latter with world.RegisterComponents()
	config := []ecs.ComponentConfig{
		{ID: TransformID, Component: TransformComponent{}},
		{ID: PhysicsID, Component: PhysicsComponent{}},
	}

	// NewWorld allocates a world and register the components
	world := ecs.NewWorld(config...)

	// World.NewEntity will add a new entity to this world
	entity := world.NewEntity()
	// World.Assign adds a list of components to the entity
	// If the entity already have the component, the Assign is ignored
	world.Assign(entity, PhysicsID, TransformID)

	// Any component registered on this entity can be retrieved using World.Component()
	// It's safe to keep this reference until the entity or the component is removed
	phys := (*PhysicsComponent)(world.Component(entity, PhysicsID))
	phys.linearAccel = Vec2D{x: 2, y: 1.5}

	// World.NewFilter creates a cache of entities that have the required components
	//
	// This solution is better than using Systems to update the entities because it's possible to
	// iterate over the filters at variable rate inside your own update function, for example,
	// the script for AI don't need to update at same frequency as physics and animations
	//
	// This filter will be automatically updated when entities or components are added/removed to the world
	filter := world.NewFilter(TransformID, PhysicsID)

	dt := float32(1.0 / 60.0)

	// filter.Entities() returns the updated list of entities that have the required components
	for _, entity := range filter.Entities() {
		// get the components for the entity
		phys := (*PhysicsComponent)(world.Component(entity, PhysicsID))
		tr := (*TransformComponent)(world.Component(entity, TransformID))

		phys.velocity.x += phys.linearAccel.x * dt
		phys.velocity.y += phys.linearAccel.y * dt

		tr.position.x += phys.velocity.x * dt
		tr.position.y += phys.velocity.y * dt

		phys.velocity.x *= 0.99
		phys.velocity.y *= 0.99
	}

	// When a filter is no longer needed, just call World.RemFilter() to remove it from the world
	// This is needed as the filters are updated when the world changes
	world.RemFilter(filter)
}

```

### Benchmarks

The benchmark folder contains the implementations of a simple test case for
performance comparison for this package as GameEngineECS, [Entitas](https://github.com/Falldot/Entitas-Go), [Ento](https://github.com/wfranczyk/ento), [Gecs](https://github.com/tutumagi/gecs), [LecsGO](https://github.com/Leopotam/go-ecs) and [EngoEngine](github.com/EngoEngine/ecs) and below are the results running on my machine.

*Notice that **EngoEngine** lets to you the responsability for keeping track of entities in the systems,
so the results may be diferent for other scenarios or usecases (to be implemented).*

**Just creation and 4 components addition to the world:**

```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go-gameengine-ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEngoEngine_100_entities_0_updates-8         	1000000000	         0.0000218 ns/op
BenchmarkEngoEngine_1000_entities_0_updates-8        	1000000000	         0.0002023 ns/op
BenchmarkEngoEngine_10000_entities_0_updates-8       	1000000000	         0.001765 ns/op
BenchmarkEngoEngine_100000_entities_0_updates-8      	1000000000	         0.02435 ns/op
BenchmarkEntitas_100_entities_0_updates-8            	1000000000	         0.0001840 ns/op
BenchmarkEntitas_1000_entities_0_updates-8           	1000000000	         0.0007055 ns/op
BenchmarkEntitas_10000_entities_0_updates-8          	1000000000	         0.008448 ns/op
BenchmarkEntitas_100000_entities_0_updates-8         	1000000000	         0.09315 ns/op
BenchmarkEnto_100_entities_0_updates-8               	1000000000	         0.0000642 ns/op
BenchmarkEnto_1000_entities_0_updates-8              	1000000000	         0.0004820 ns/op
BenchmarkEnto_10000_entities_0_updates-8             	1000000000	         0.003454 ns/op
BenchmarkEnto_100000_entities_0_updates-8            	1000000000	         0.04797 ns/op
BenchmarkGameEngineECS_100_entities_0_updates-8      	1000000000	         0.0000849 ns/op
BenchmarkGameEngineECS_1000_entities_0_updates-8     	1000000000	         0.0001953 ns/op
BenchmarkGameEngineECS_10000_entities_0_updates-8    	1000000000	         0.001686 ns/op
BenchmarkGameEngineECS_100000_entities_0_updates-8   	1000000000	         0.02034 ns/op
BenchmarkGecs_100_entities_0_updates-8               	1000000000	         0.0007248 ns/op
BenchmarkGecs_1000_entities_0_updates-8              	1000000000	         0.0003266 ns/op
BenchmarkGecs_10000_entities_0_updates-8             	1000000000	         0.002361 ns/op
BenchmarkGecs_100000_entities_0_updates-8            	1000000000	         0.02372 ns/op
BenchmarkLecsGO_100_entities_0_updates-8             	1000000000	         0.0000705 ns/op
BenchmarkLecsGO_1000_entities_0_updates-8            	1000000000	         0.0004838 ns/op
BenchmarkLecsGO_10000_entities_0_updates-8           	1000000000	         0.006037 ns/op
BenchmarkLecsGO_100000_entities_0_updates-8          	1000000000	         0.07106 ns/op
PASS
ok  	github.com/marioolofo/go-gameengine-ecs/benchmark	2.654s
```
![0 iterations graph](./benchmark/results/0updates.png)

#### Iteration time for 100~10,000 entities:

**100 iterations:**

```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go-gameengine-ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEngoEngine_100_entities_100_updates-8         	1000000000	         0.0000845 ns/op
BenchmarkEngoEngine_1000_entities_100_updates-8        	1000000000	         0.0009191 ns/op
BenchmarkEngoEngine_10000_entities_100_updates-8       	1000000000	         0.009456 ns/op
BenchmarkEngoEngine_100000_entities_100_updates-8      	1000000000	         0.1757 ns/op
BenchmarkEntitas_100_entities_100_updates-8            	1000000000	         0.0002135 ns/op
BenchmarkEntitas_1000_entities_100_updates-8           	1000000000	         0.002982 ns/op
BenchmarkEntitas_10000_entities_100_updates-8          	1000000000	         0.03583 ns/op
BenchmarkEntitas_100000_entities_100_updates-8         	       1	1126714665 ns/op
BenchmarkEnto_100_entities_100_updates-8               	1000000000	         0.001355 ns/op
BenchmarkEnto_1000_entities_100_updates-8              	1000000000	         0.01245 ns/op
BenchmarkEnto_10000_entities_100_updates-8             	1000000000	         0.1271 ns/op
BenchmarkEnto_100000_entities_100_updates-8            	       1	1325536011 ns/op
BenchmarkGameEngineECS_100_entities_100_updates-8      	1000000000	         0.0002114 ns/op
BenchmarkGameEngineECS_1000_entities_100_updates-8     	1000000000	         0.001107 ns/op
BenchmarkGameEngineECS_10000_entities_100_updates-8    	1000000000	         0.01083 ns/op
BenchmarkGameEngineECS_100000_entities_100_updates-8   	1000000000	         0.1112 ns/op
BenchmarkGecs_100_entities_100_updates-8               	1000000000	         0.0007154 ns/op
BenchmarkGecs_1000_entities_100_updates-8              	1000000000	         0.007240 ns/op
BenchmarkGecs_10000_entities_100_updates-8             	1000000000	         0.07199 ns/op
BenchmarkGecs_100000_entities_100_updates-8            	1000000000	         0.7301 ns/op
BenchmarkLecsGO_100_entities_100_updates-8             	1000000000	         0.0000740 ns/op
BenchmarkLecsGO_1000_entities_100_updates-8            	1000000000	         0.0007187 ns/op
BenchmarkLecsGO_10000_entities_100_updates-8           	1000000000	         0.009319 ns/op
BenchmarkLecsGO_100000_entities_100_updates-8          	1000000000	         0.1292 ns/op
PASS
ok  	github.com/marioolofo/go-gameengine-ecs/benchmark	42.865s
```
![100 iterations graph](./benchmark/results/100updates.png)

**1,000 iterations:**

```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go-gameengine-ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEngoEngine_100_entities_1000_updates-8         	1000000000	         0.0006254 ns/op
BenchmarkEngoEngine_1000_entities_1000_updates-8        	1000000000	         0.007021 ns/op
BenchmarkEngoEngine_10000_entities_1000_updates-8       	1000000000	         0.07600 ns/op
BenchmarkEngoEngine_100000_entities_1000_updates-8      	       1	1349145574 ns/op
BenchmarkEntitas_100_entities_1000_updates-8            	1000000000	         0.001256 ns/op
BenchmarkEntitas_1000_entities_1000_updates-8           	1000000000	         0.02259 ns/op
BenchmarkEntitas_10000_entities_1000_updates-8          	1000000000	         0.2912 ns/op
BenchmarkEntitas_100000_entities_1000_updates-8         	       1	10328067642 ns/op
BenchmarkEnto_100_entities_1000_updates-8               	1000000000	         0.01217 ns/op
BenchmarkEnto_1000_entities_1000_updates-8              	1000000000	         0.1232 ns/op
BenchmarkEnto_10000_entities_1000_updates-8             	       1	1236405407 ns/op
BenchmarkEnto_100000_entities_1000_updates-8            	       1	12721365771 ns/op
BenchmarkGameEngineECS_100_entities_1000_updates-8      	1000000000	         0.001022 ns/op
BenchmarkGameEngineECS_1000_entities_1000_updates-8     	1000000000	         0.009422 ns/op
BenchmarkGameEngineECS_10000_entities_1000_updates-8    	1000000000	         0.09406 ns/op
BenchmarkGameEngineECS_100000_entities_1000_updates-8   	1000000000	         0.9446 ns/op
BenchmarkGecs_100_entities_1000_updates-8               	1000000000	         0.007726 ns/op
BenchmarkGecs_1000_entities_1000_updates-8              	1000000000	         0.06897 ns/op
BenchmarkGecs_10000_entities_1000_updates-8             	1000000000	         0.7034 ns/op
BenchmarkGecs_100000_entities_1000_updates-8            	       1	7117594882 ns/op
BenchmarkLecsGO_100_entities_1000_updates-8             	1000000000	         0.0005196 ns/op
BenchmarkLecsGO_1000_entities_1000_updates-8            	1000000000	         0.003643 ns/op
BenchmarkLecsGO_10000_entities_1000_updates-8           	1000000000	         0.03769 ns/op
BenchmarkLecsGO_100000_entities_1000_updates-8          	1000000000	         0.6236 ns/op
PASS
ok  	github.com/marioolofo/go-gameengine-ecs/benchmark	177.623s
```
![1,000 iterations graph](./benchmark/results/1000updates.png)

**10,000 iterations:**

```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go-gameengine-ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEngoEngine_100_entities_10000_updates-8         	1000000000	         0.005653 ns/op
BenchmarkEngoEngine_1000_entities_10000_updates-8        	1000000000	         0.06794 ns/op
BenchmarkEngoEngine_10000_entities_10000_updates-8       	1000000000	         0.7393 ns/op
BenchmarkEngoEngine_100000_entities_10000_updates-8      	       1	13142254965 ns/op
BenchmarkEntitas_100_entities_10000_updates-8            	1000000000	         0.01188 ns/op
BenchmarkEntitas_1000_entities_10000_updates-8           	1000000000	         0.2206 ns/op
BenchmarkEntitas_10000_entities_10000_updates-8          	       1	2856435193 ns/op
BenchmarkEntitas_100000_entities_10000_updates-8         	       1	105500478112 ns/op
BenchmarkEnto_100_entities_10000_updates-8               	1000000000	         0.1234 ns/op
BenchmarkEnto_1000_entities_10000_updates-8              	       1	1215646893 ns/op
BenchmarkEnto_10000_entities_10000_updates-8             	       1	12336209652 ns/op
BenchmarkEnto_100000_entities_10000_updates-8            	       1	127410246972 ns/op
BenchmarkGameEngineECS_100_entities_10000_updates-8      	1000000000	         0.009515 ns/op
BenchmarkGameEngineECS_1000_entities_10000_updates-8     	1000000000	         0.09226 ns/op
BenchmarkGameEngineECS_10000_entities_10000_updates-8    	1000000000	         0.9229 ns/op
BenchmarkGameEngineECS_100000_entities_10000_updates-8   	       1	9280284220 ns/op
BenchmarkGecs_100_entities_10000_updates-8               	1000000000	         0.07238 ns/op
BenchmarkGecs_1000_entities_10000_updates-8              	      34	  31453421 ns/op
BenchmarkGecs_10000_entities_10000_updates-8             	       1	7598556343 ns/op
BenchmarkGecs_100000_entities_10000_updates-8            	       1	70832011441 ns/op
BenchmarkLecsGO_100_entities_10000_updates-8             	1000000000	         0.01323 ns/op
BenchmarkLecsGO_1000_entities_10000_updates-8            	1000000000	         0.1331 ns/op
BenchmarkLecsGO_10000_entities_10000_updates-8           	       1	1361573818 ns/op
BenchmarkLecsGO_100000_entities_10000_updates-8          	       1	15729605049 ns/op
PASS
ok  	github.com/marioolofo/go-gameengine-ecs/benchmark	492.671s
```
![10,000 iterations graph](./benchmark/results/10000updates.png)

## License

This project is distributed under the MIT licence.

```
MIT License

Copyright (c) 2022 Mario Olofo <mario.olofo@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.


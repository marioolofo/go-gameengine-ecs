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

From [simple.go](./examples/simple/simple.go):

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
performance comparison for this package as GameEngineECS, [Entitas](https://github.com/Falldot/Entitas-Go), [Ento](https://github.com/wfranczyk/ento), [Gecs](https://github.com/tutumagi/gecs) and [LecsGO](https://github.com/Leopotam/go-ecs) and below are the results running on my machine:

**Just creation and 4 components addition to the world:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8              1000000000               0.0000043 ns/op
BenchmarkEnto-8                 1000000000               0.0000147 ns/op
BenchmarkGecs-8                 1000000000               0.0000694 ns/op
BenchmarkLecsGO-8               1000000000               0.0000146 ns/op
BenchmarkGameEngineECS-8        1000000000               0.0000092 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       0.008s
```

#### Iteration time for 100 entities:
**1000 iterations:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8              1000000000               0.003533 ns/op
BenchmarkEnto-8                 1000000000               0.02393 ns/op
BenchmarkGecs-8                 1000000000               0.01464 ns/op
BenchmarkLecsGO-8               1000000000               0.0006838 ns/op
BenchmarkGameEngineECS-8        1000000000               0.002037 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       0.308s
```

**10000 iterations:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8              1000000000               0.02744 ns/op
BenchmarkEnto-8                 1000000000               0.2404 ns/op
BenchmarkGecs-8                 1000000000               0.1638 ns/op
BenchmarkLecsGO-8               1000000000               0.02645 ns/op
BenchmarkGameEngineECS-8        1000000000               0.01831 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       5.679s
```

**100000 iterations:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8              1000000000           0.2639 ns/op
BenchmarkEnto-8                         1        2380309375 ns/op
BenchmarkGecs-8                         1        1749332537 ns/op
BenchmarkLecsGO-8                       1        1366320329 ns/op
BenchmarkGameEngineECS-8        1000000000           0.1825 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       11.826s
```

#### 100 iterations for x entities allocated:
**10000 entities:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8              1000000000               0.1011 ns/op
BenchmarkEnto-8                 1000000000               0.2553 ns/op
BenchmarkGecs-8                 1000000000               0.1471 ns/op
BenchmarkLecsGO-8               1000000000               0.01919 ns/op
BenchmarkGameEngineECS-8        1000000000               0.02215 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       6.777s
```

**50000 entities:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8                     1        1064971463 ns/op
BenchmarkEnto-8                        1        1303889925 ns/op
BenchmarkGecs-8                 1000000000               0.7330 ns/op
BenchmarkLecsGO-8               1000000000               0.1251 ns/op
BenchmarkGameEngineECS-8        1000000000               0.1117 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       37.989s
```

**100000 iterations:**
```sh
goos: linux
goarch: amd64
pkg: github.com/marioolofo/go/gameengine/ecs/benchmark
cpu: Intel(R) Core(TM) i5-8300H CPU @ 2.30GHz
BenchmarkEntitas-8                     1        2276978192 ns/op
BenchmarkEnto-8                        1        2609283178 ns/op
BenchmarkGecs-8                        1        1448020979 ns/op
BenchmarkLecsGO-8               1000000000               0.2732 ns/op
BenchmarkGameEngineECS-8        1000000000               0.2230 ns/op
PASS
ok      github.com/marioolofo/go/gameengine/ecs/benchmark       13.763s
```


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

```

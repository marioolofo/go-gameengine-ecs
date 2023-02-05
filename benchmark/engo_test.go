package ecs_benchmark

import (
	"fmt"
	"testing"

	"github.com/EngoEngine/ecs"
)

type EditorOption struct {
	ecs.BasicEntity
	ui     UIDesign
	script Script
}

type Object2D struct {
	ecs.BasicEntity
	transform Transform2D
	physics   Physics2D
}

func (s *Script) GetScript() *Script {
	return s
}

func (ui *UIDesign) GetUIDesign() *UIDesign {
	return ui
}

func (tr *Transform2D) GetTransform2D() *Transform2D {
	return tr
}

func (p *Physics2D) GetPhysics2D() *Physics2D {
	return p
}

type ScriptFace interface {
	GetScript() *Script
}

type UIDesignFace interface {
	GetUIDesign() *UIDesign
}

type Physics2DFace interface {
	GetPhysics2D() *Physics2D
}

type Transform2DFace interface {
	GetTransform2D() *Transform2D
}

type Editorable interface {
	ecs.BasicFace
	UIDesignFace
	ScriptFace
}

type Physicsable interface {
	ecs.BasicFace
	Physics2DFace
	Transform2DFace
}

type editorSystemEntity struct {
	ui     *UIDesign
	script *Script
}

type EditorSystem struct {
	entities map[uint64]editorSystemEntity
}

func (editor *EditorSystem) New(w *ecs.World) {
	editor.entities = make(map[uint64]editorSystemEntity)
}

func (editor *EditorSystem) Add(e *ecs.BasicEntity, ui *UIDesign, script *Script) {
	editor.entities[e.ID()] = editorSystemEntity{ui, script}
}

func (editor *EditorSystem) AddByInterface(o ecs.Identifier) {
	fmt.Println("Add by interface editor")
	obj := o.(Editorable)
	editor.Add(obj.GetBasicEntity(), obj.GetUIDesign(), obj.GetScript())
}

func (editor *EditorSystem) Update(dt float32) {
}

func (editor *EditorSystem) Remove(e ecs.BasicEntity) {
	delete(editor.entities, e.ID())
}

type objectSystemEntity struct {
	physics   *Physics2D
	transform *Transform2D
}

type ObjectSystem struct {
	entities map[uint64]objectSystemEntity
}

func (o *ObjectSystem) New(w *ecs.World) {
	o.entities = make(map[uint64]objectSystemEntity)
}

func (o *ObjectSystem) Add(e *ecs.BasicEntity, physics *Physics2D, transform *Transform2D) {
	o.entities[e.ID()] = objectSystemEntity{physics, transform}
}

func (o *ObjectSystem) AddByInterface(i ecs.Identifier) {
	obj := i.(Physicsable)
	o.Add(obj.GetBasicEntity(), obj.GetPhysics2D(), obj.GetTransform2D())
}

func (o *ObjectSystem) Update(dt float32) {
	for _, entity := range o.entities {
		phys := entity.physics
		tr := entity.transform

		phys.velocity.x += phys.linearAccel.x * dt
		phys.velocity.y += phys.linearAccel.y * dt

		tr.position.x += phys.velocity.x * dt
		tr.position.y += phys.velocity.y * dt

		phys.velocity.x *= 0.99
		phys.velocity.y *= 0.99
	}
}

func (o *ObjectSystem) Remove(e ecs.BasicEntity) {
	delete(o.entities, e.ID())
}

func EngoEngineBench(b *testing.B, entityCount, updateCount int) {

	world := ecs.World{}

	world.AddSystem(&EditorSystem{})
	world.AddSystem(&ObjectSystem{})

	var editorSys *EditorSystem
	var objSys *ObjectSystem

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *EditorSystem:
			{
				editorSys = sys
			}
		case *ObjectSystem:
			{
				objSys = sys
			}
		}
	}

	for i := 0; i < entityCount/2; i++ {
		e1 := EditorOption{BasicEntity: ecs.NewBasic()}
		e1.ui.name = fmt.Sprint("entity_", i)
		editorSys.Add(&e1.BasicEntity, &e1.ui, &e1.script)

		e2 := Object2D{BasicEntity: ecs.NewBasic()}
		e2.physics.linearAccel = Vec2D{x: 2, y: 1.5}
		objSys.Add(&e2.BasicEntity, &e2.physics, &e2.transform)
	}

	for i := 0; i < updateCount; i++ {
		world.Update(dt)
	}
}

// 0 updates

func BenchmarkEngoEngine_100_entities_0_updates(b *testing.B) {
	EngoEngineBench(b, 100, 0)
}

func BenchmarkEngoEngine_1000_entities_0_updates(b *testing.B) {
	EngoEngineBench(b, 1000, 0)
}

func BenchmarkEngoEngine_10000_entities_0_updates(b *testing.B) {
	EngoEngineBench(b, 10000, 0)
}

func BenchmarkEngoEngine_100000_entities_0_updates(b *testing.B) {
	EngoEngineBench(b, 100000, 0)
}

// 100 updates

func BenchmarkEngoEngine_100_entities_100_updates(b *testing.B) {
	EngoEngineBench(b, 100, 100)
}

func BenchmarkEngoEngine_1000_entities_100_updates(b *testing.B) {
	EngoEngineBench(b, 1000, 100)
}

func BenchmarkEngoEngine_10000_entities_100_updates(b *testing.B) {
	EngoEngineBench(b, 10000, 100)
}

func BenchmarkEngoEngine_100000_entities_100_updates(b *testing.B) {
	EngoEngineBench(b, 100000, 100)
}

// 1000 updates

func BenchmarkEngoEngine_100_entities_1000_updates(b *testing.B) {
	EngoEngineBench(b, 100, 1000)
}

func BenchmarkEngoEngine_1000_entities_1000_updates(b *testing.B) {
	EngoEngineBench(b, 1000, 1000)
}

func BenchmarkEngoEngine_10000_entities_1000_updates(b *testing.B) {
	EngoEngineBench(b, 10000, 1000)
}

func BenchmarkEngoEngine_100000_entities_1000_updates(b *testing.B) {
	EngoEngineBench(b, 100000, 1000)
}

// 10000 updates

func BenchmarkEngoEngine_100_entities_10000_updates(b *testing.B) {
	EngoEngineBench(b, 100, 10000)
}

func BenchmarkEngoEngine_1000_entities_10000_updates(b *testing.B) {
	EngoEngineBench(b, 1000, 10000)
}

func BenchmarkEngoEngine_10000_entities_10000_updates(b *testing.B) {
	EngoEngineBench(b, 10000, 10000)
}

func BenchmarkEngoEngine_100000_entities_10000_updates(b *testing.B) {
	EngoEngineBench(b, 100000, 10000)
}

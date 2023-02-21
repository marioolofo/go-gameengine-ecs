package ecs

type EntityPool interface {
	New() EntityID
	NewComponent() EntityID
	Recycle(e EntityID) bool
	IsAlive(e EntityID) bool
}

const (
	EntityPoolInitialCapacity = 1024 * 10
)

type entityPool struct {
	entities  []EntityID
	next      uint64
	available uint64
}

func NewEntityPool(initialCap uint) EntityPool {
	if initialCap == 0 {
		initialCap = EntityPoolInitialCapacity
	}
	ep := &entityPool{
		entities:  make([]EntityID, 0, initialCap+1),
		next:      0,
		available: 0,
	}

	ep.New()

	return ep
}

func (e *entityPool) New() EntityID {
	if e.available > 0 {
		e.available--
		index := e.next
		entity := e.entities[index]
		e.next = entity.ID()
		e.entities[index] = entity.SetID(index)
		return e.entities[index]
	}

	entity := MakeEntity(uint64(len(e.entities)), 0)
	e.entities = append(e.entities, entity)
	return entity
}

func (e *entityPool) NewComponent() EntityID {
	entity := e.New()
	index := entity.ID()
	e.entities[index] = entity
	return entity
}

func (e *entityPool) Recycle(entity EntityID) bool {
	if !e.IsAlive(entity) {
		return false
	}
	index := entity.ID()

	e.available++
	e.entities[index] = MakeEntity(e.next, entity.Gen()+1)
	e.next = index

	return true
}

func (e entityPool) IsAlive(entity EntityID) bool {
	index := entity.ID()
	if index == 0 ||
		index >= uint64(len(e.entities)) {
		return false
	}
	return e.entities[entity.ID()] == entity.WithoutFlags()
}

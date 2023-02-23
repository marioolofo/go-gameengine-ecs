package ecs

/*
	EntityPool defines the interface for the entity ID manager.
	It uses an implicit linked list to keep track of all recycled IDs without an additional buffer.
	When a recycled ID is returned, it have its generation bits incremented to differentiate it from
	the older one.

	New returns a new EntityID

	Recycle puts the EntityID in the recycle list for reuse. If the entity is not alive, returns false and do nothing.

	IsAlive returns true if the EntityID is alive in the pool
*/
type EntityPool interface {
	New() EntityID
	Recycle(e EntityID) bool
	IsAlive(e EntityID) bool
}

const (
	EntityPoolInitialCapacity = 1024 * 10 // default initial buffer size
)

type entityPool struct {
	entities  []EntityID
	next      uint64
	available uint64
}

/*
	NewEntityPool returns a implementation of EntityPool with initialCap as capacity.
	If the initialCap == 0, the initial capcity is set to EntityPoolInitialCapacity
*/
func NewEntityPool(initialCap uint) EntityPool {
	if initialCap == 0 {
		initialCap = EntityPoolInitialCapacity
	}
	ep := &entityPool{
		entities:  make([]EntityID, 1, initialCap+1),
		next:      0,
		available: 0,
	}

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

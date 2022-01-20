package ecs

import "math"

// ID defines the format for the components identifier
type ID uint32
// Entity defines the format of the entity identifier
type Entity = ID
// Index defines the integer size to use for sparse arrays
type Index uint32

// InvalidIndex defines the value for invalid indices
const InvalidIndex = math.MaxUint32

// Configuration for the size of internal arrays
// Change this values only before any call to NewWorld()
var (
	InitialEntityCapacity = 1024 * 10
 	InitialEntityRecycleCapacity = 256
 	InitialSparseArrayCapacity = 1024
 	InitialMemoryPoolCapacityShift = 10
 	InitialMemoryPoolCapacity = 1<<InitialMemoryPoolCapacityShift
)

package ecs

import (
	"unsafe"
)

// MemoryPool is the interface that controls the allocation and access of blocks of memory
// The MemoryPool allocate pages of memory with total size defined by itemSizeAligned * InitialMemoryPoolCapacity,
// this has the advantage of never risk to lose pointer references due to some slice growth
//
// Alloc reuse or allocate a block of memory and returns the unsafe.Pointer and Index for that block
// The contents of the new block is always zero
// It's safe to keep and use the pointer until a call to Free(Index) is done
//
// Free recycles the block identified by Index
//
// Get returns the unsafe.Pointer for the memory block identified by Index or nil if the index is invalid
//
// Set fill the memory block at Index with the contents of interface{}
//
// Zero fills the memory block at Index with zeros
//
// Reset discards all the blocks and sets the MemoryPool to its initial state
//
// Stats collects statistics from the MemoryPool
type MemoryPool interface {
	Alloc() (unsafe.Pointer, Index)
	Free(Index)
	Get(Index) unsafe.Pointer
	Set(Index, interface{}) bool
	Zero(Index)
	Reset()
	Stats() MemoryPoolStats
}

// MemoryPoolStats is the runtime information of the MemoryPool
type MemoryPoolStats struct {
	ID                   // MemoryPool ID
	InUse           uint // Total components in use
	Recycled        uint // Total components in recycle list
	SizeInBytes     uint // Allocation size in Bytes for every instance of this component
	Alignment       uint // This component is laid in memory in multiples of this value
	PageSizeInBytes uint // Allocation size in Bytes for a page of components
	TotalPages      uint // Total pages allocated
}

// MemoryPoolConfg defines the layout of the MemoryPool
type MemoryPoolConfg struct {
	id             ID          // id identifies the MemoryPool
	obj            interface{} // obj is an instance of the object used as template for the block size and alignment
	forceAlignment int         // forceAlignment force all the blocks to have this alignment if forceAlignment > 0
}

// NewMemoryPool returns a new MemoryPool for a given config
func NewMemoryPool(id ID, objectRef interface{}, forceAlignment int) MemoryPool {
	return NewMemoryPoolReflect(id, objectRef)
}

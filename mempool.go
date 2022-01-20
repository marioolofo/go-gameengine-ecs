package ecs

import (
	"log"
	"reflect"
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
type MemoryPool interface {
	Alloc() (unsafe.Pointer, Index)
	Free(Index)
	Get(Index) unsafe.Pointer
	Set(Index, interface{})
	Zero(Index)
	Reset()
}

type memoryPage struct {
	used int
	buffer []byte
}

type memPool struct {
	id ID
	itemSize uintptr
	itemSizeAligned uintptr
	itemMask uintptr
	recycle []uintptr
	pages []memoryPage
}

// MemoryPoolConfg defines the layout of the MemoryPool
type MemoryPoolConfg struct {
	id ID              // id identifies the MemoryPool
	obj interface{}    // obj is an instance of the object used as template for the block size and alignment
	forceAlignment int // forceAlignment force all the blocks to have this alignment if forceAlignment > 0
}

// NewMemoryPool returns a new MemoryPool for a given config
func NewMemoryPool(id ID, objectRef interface{}, forceAlignment int) MemoryPool {
	typeOf := reflect.TypeOf(objectRef)
	size := typeOf.Size()
	align := typeOf.Align()

	if forceAlignment != 0 && forceAlignment&(forceAlignment - 1) != 0 {
		log.Printf("[Sys] forceAlignment not power of two (%d)\n", forceAlignment)
	} else if forceAlignment > align {
		align = forceAlignment
	}

	dataSizeAligned := (size + uintptr(align) - 1) & ^uintptr(align - 1)
	mask := ^(dataSizeAligned-1)

	s := &memPool{
		id,
		size,
		dataSizeAligned,
		mask,
		make([]uintptr, 0),
		make([]memoryPage, 0),
	}

	return s
}

func (s *memPool) Alloc() (unsafe.Pointer, Index) {
	if len(s.pages) == 0 || s.pages[len(s.pages)-1].used == InitialMemoryPoolCapacity {
		s.newMemoryPage()
	}

	var index uintptr

	if len(s.recycle) > 0 {
		index = uintptr(s.recycle[len(s.recycle)-1])
		s.recycle = s.recycle[:len(s.recycle)-1]
		s.Zero(Index(index))
	} else {
		last := len(s.pages) - 1
		index = uintptr(last<<InitialMemoryPoolCapacityShift + s.pages[last].used)
		s.pages[last].used++
	}

	return s.Get(Index(index)), Index(index)
}

func (s *memPool) Free(index Index) {
	s.recycle = append(s.recycle, uintptr(index))
}

func (s *memPool) Get(index Index) unsafe.Pointer {
	if int(index) >= len(s.pages) << InitialMemoryPoolCapacityShift {
		return nil
	}
	ind := int(index) >> InitialMemoryPoolCapacityShift
	page := unsafe.Pointer(&s.pages[ind].buffer[0])
	alignment := s.itemSizeAligned - (uintptr(page) & (s.itemSizeAligned-1))
	offset := alignment + (uintptr(index) & uintptr(InitialMemoryPoolCapacity - 1)) * s.itemSizeAligned
	return unsafe.Add(page, offset)
}

func (s *memPool) Set(index Index, value interface{}) {
	dst := s.Get(index)
	src := reflect.ValueOf(value).UnsafePointer()
	for i := uintptr(0); i < s.itemSize; i++ {
		*(*byte)(dst) = *(*byte)(src)
		dst = unsafe.Add(dst, 1)
		src = unsafe.Add(src, 1)
	}
}

func (s *memPool) Zero(index Index) {
	dst := s.Get(index)
	for i := uintptr(0); i < s.itemSize; i++ {
		*(*byte)(dst) = 0
		dst = unsafe.Add(dst, 1)
	}
}

func (s *memPool) Reset() {
	s.recycle = make([]uintptr, 0)
	s.pages = nil
}

func (s *memPool) newMemoryPage() {
	bufferSize := s.itemSizeAligned + s.itemSizeAligned<<InitialMemoryPoolCapacityShift
	buffer := make([]byte, bufferSize)

	s.pages = append(s.pages, memoryPage{
		0,
		buffer,
	})
}

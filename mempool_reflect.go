package ecs

import (
	"math"
	"reflect"
	"unsafe"
)

type memoryPageReflect struct {
	used          int
	buffer        reflect.Value
	bufferAddress unsafe.Pointer
}

type memoryPoolReflect struct {
	id       ID
	typeOf   reflect.Type
	itemSize uintptr
	recycle  []uintptr
	pages    []memoryPageReflect
}

// NewMemoryPoolReflect returns an implementation for MemoryPool that uses reflect to
// create and manage the arrays of structures.
// It's safe and work correctly with the GC, but it's a little slower than other implementations
func NewMemoryPoolReflect(id ID, objectRef interface{}) MemoryPool {
	typeOf := reflect.TypeOf(objectRef)
	align := uintptr(typeOf.Align())
	size := typeOf.Size()

	size = (size + (align - 1)) / align * align

	return &memoryPoolReflect{
		id,
		typeOf,
		size,
		make([]uintptr, 0),
		make([]memoryPageReflect, 0),
	}
}

func (s *memoryPoolReflect) Alloc() (unsafe.Pointer, Index) {
	if len(s.pages) == 0 || s.pages[len(s.pages)-1].used == InitialMemoryPoolCapacity {
		s.newMemoryPage()
	}

	var index uintptr

	if len(s.recycle) > 0 {
		index = s.recycle[len(s.recycle)-1]
		s.recycle = s.recycle[:len(s.recycle)-1]
		s.Zero(Index(index))
	} else {
		last := len(s.pages) - 1
		index = uintptr(last<<InitialMemoryPoolCapacityShift + s.pages[last].used)
		s.pages[last].used++
	}

	return s.Get(Index(index)), Index(index)
}

func (s *memoryPoolReflect) Free(index Index) {
	s.recycle = append(s.recycle, uintptr(index))
}

func (s *memoryPoolReflect) Get(index Index) unsafe.Pointer {
	if int(index) >= len(s.pages)<<InitialMemoryPoolCapacityShift {
		return nil
	}
	ind := int(index) >> InitialMemoryPoolCapacityShift
	offset := int(index) & (InitialMemoryPoolCapacity - 1)
	ptr := unsafe.Add(s.pages[ind].bufferAddress, uintptr(offset)*s.itemSize)
	return unsafe.Pointer(ptr)
}

func (s *memoryPoolReflect) Set(index Index, value interface{}) bool {
	rValue := reflect.ValueOf(value)
	dst := s.Get(index)
	var src unsafe.Pointer
	size := s.itemSize

	if rValue.Kind() != reflect.Pointer {
		return false
	}
	src = rValue.UnsafePointer()

	dstSlice := (*[math.MaxInt32]byte)(dst)[:size:size]
	srcSlice := (*[math.MaxInt32]byte)(src)[:size:size]

	return copy(dstSlice, srcSlice) == int(s.itemSize)
}

func (s *memoryPoolReflect) Zero(index Index) {
	dst := s.Get(index)

	for i := uintptr(0); i < s.itemSize; i++ {
		*(*byte)(dst) = 0
		dst = unsafe.Add(dst, 1)
	}
}

func (s *memoryPoolReflect) Reset() {
	s.recycle = make([]uintptr, 0)
	s.pages = nil
}

func (s *memoryPoolReflect) Stats() MemoryPoolStats {
	totalPages := uint(len(s.pages))
	itemSize := uint(s.itemSize)
	recycled := uint(len(s.recycle))
	lastPage := uint(0)
	lastItem := uint(0)

	if totalPages > 0 {
		lastPage = totalPages - 1
		lastItem = lastPage<<InitialMemoryPoolCapacityShift + uint(s.pages[lastPage].used)
	}

	bufferSize := s.itemSize << InitialMemoryPoolCapacityShift

	return MemoryPoolStats{
		ID:              s.id,
		InUse:           lastItem - recycled,
		Recycled:        uint(len(s.recycle)),
		SizeInBytes:     itemSize,
		Alignment:       uint(s.itemSize),
		PageSizeInBytes: uint(bufferSize),
		TotalPages:      totalPages,
	}
}

func (s *memoryPoolReflect) newMemoryPage() {
	buffer := reflect.New(reflect.ArrayOf(InitialMemoryPoolCapacity, s.typeOf)).Elem()
	s.pages = append(s.pages, memoryPageReflect{
		0,
		buffer,
		buffer.Addr().UnsafePointer(),
	})
}

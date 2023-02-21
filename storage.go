// go:build -gcflags=-B
package ecs

import (
	"reflect"
	"unsafe"
)

// Storage is the interface that controls the allocation and access of blocks of memory.
//
// Get returns the unsafe.Pointer for the memory block identified by Index or nil if the index is invalid
//
// Set fill the memory block at Index with the contents of interface{}
//
// Remove removes the entry at index
//
// Reset discards all the blocks and sets the Storage to its initial state
//
// Stats collects statistics from the Storage
type Storage interface {
	Get(uint) unsafe.Pointer
	Set(uint, interface{}) bool
	Copy(uint, unsafe.Pointer)
	Shrink(uint)
	Expand(uint)
	Reset()
	Stats() StorageStats
}

// StorageStats is the runtime information of a specific Storage
type StorageStats struct {
	Type     reflect.Type
	ItemSize uint // size in Bytes for every instance of the item
	Cap      uint // current item capacity
}

const (
	StorageBufferIncrementBy = 10000
)

type storage[T any] struct {
	increment uint
	buffer    []T
	bufferPtr unsafe.Pointer
}

// NewStorage returns an implementation for Storage
// If increment is zero, StorageBufferIncrementBy will be used
func NewStorage[T any](initialLen, increment uint) Storage {
	if increment == 0 {
		increment = StorageBufferIncrementBy
	}

	buffer := make([]T, initialLen, initialLen)
	return &storage[T]{
		increment,
		buffer,
		unsafe.Pointer(&buffer[0]),
	}
}

func (s *storage[T]) Get(index uint) unsafe.Pointer {
	var t T
	return unsafe.Add(s.bufferPtr, unsafe.Sizeof(t)*uintptr(index))
}

func (s *storage[T]) Set(index uint, value interface{}) bool {
	if int(index) >= len(s.buffer) {
		return false
	}
	v, ok := value.(*T)
	if ok {
		ptr := (*T)(s.Get(index))
		*ptr = *v
	}
	return ok
}

func (s *storage[T]) Copy(index uint, ptr unsafe.Pointer) {
	to := (*T)(s.Get(index))
	from := (*T)(ptr)
	*to = *from
}

func (s *storage[T]) Shrink(to uint) {
	if to < uint(len(s.buffer)) {
		prev := s.buffer[0:int(to)]
		s.buffer = make([]T, int(to))
		copy(s.buffer, prev)
	}
}

func (s *storage[T]) Expand(to uint) {
	s.ensureCap(to)
}

func (s *storage[T]) Reset() {
	s.buffer = make([]T, 0, 0)
	s.bufferPtr = unsafe.Pointer(nil)
}

func (s storage[T]) Stats() StorageStats {
	var t T
	typeOf := reflect.TypeOf(t)

	return StorageStats{
		typeOf,
		uint(typeOf.Size()),
		uint(len(s.buffer)),
	}
}

func (s *storage[T]) ensureCap(size uint) {
	if size >= uint(cap(s.buffer)) {
		prevBuffer := s.buffer
		s.buffer = make([]T, int(size+s.increment), int(size+s.increment))
		s.bufferPtr = unsafe.Pointer(&s.buffer[0])
		copy(s.buffer, prevBuffer)
	}
}

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
	StorageBufferIncrementBy = 1000
)

type storage[T any] struct {
	increment uint
	buffer    []T
}

// NewStorage returns an implementation for Storage
// If increment is zero, StorageBufferIncrementBy will be used
func NewStorage[T any](initialLen, increment uint) Storage {
	if increment == 0 {
		increment = StorageBufferIncrementBy
	}

	return &storage[T]{
		increment,
		make([]T, initialLen, initialLen),
	}
}

func (s *storage[T]) Get(index uint) unsafe.Pointer {
	if int(index) >= cap(s.buffer) {
		s.ensureCap(index)
	}
	return unsafe.Pointer(&s.buffer[index])
}

func (s *storage[T]) Set(index uint, value interface{}) bool {
	v, ok := value.(*T)
	if index >= uint(cap(s.buffer)) || !ok {
		return false
	}
	s.buffer[index] = *v
	return true
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
	if to > uint(len(s.buffer)) {
		prev := s.buffer
		s.buffer = make([]T, int(to))
		copy(s.buffer, prev)
	}
}

func (s *storage[T]) Reset() {
	s.buffer = make([]T, 0, 0)
}

func (s *storage[T]) Stats() StorageStats {
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
		copy(s.buffer, prevBuffer)
	}
}

package ecs

import (
	"math"
	"reflect"
	"unsafe"
)

type storageReflect struct {
	buffer            reflect.Value
	bufferAddress     unsafe.Pointer
	typeOf            reflect.Type
	itemSize          uintptr
	capacityIncrement uint32
}

/*
 NewStorageReflect creates a new Storage using reflection.

 This implementation is a bit slower than the generic version, as it have
 to detect in runtime the size of the item to index the address
*/
func NewStorageReflect(ref interface{}, initialLen, increment uint) Storage {
	if increment == 0 {
		increment = StorageBufferIncrementBy
	}

	tp := reflect.Indirect(reflect.ValueOf(ref)).Type()
	size := tp.Size()
	align := uintptr(tp.Align())
	size = (size + (align - 1)) / align * align

	buffer := reflect.New(reflect.ArrayOf(int(initialLen), tp)).Elem()
	return &storageReflect{
		buffer:            buffer,
		bufferAddress:     buffer.Addr().UnsafePointer(),
		typeOf:            tp,
		itemSize:          size,
		capacityIncrement: uint32(increment),
	}
}

func (s *storageReflect) Get(index uint) unsafe.Pointer {
	return unsafe.Add(s.bufferAddress, uintptr(index)*s.itemSize)
}

func (s *storageReflect) Set(index uint, value interface{}) bool {
	if s.itemSize == 0 || index >= uint(s.buffer.Cap()) {
		return false
	}

	dst := s.Get(index)

	rValue := reflect.ValueOf(value)

	var src unsafe.Pointer
	size := s.itemSize

	src = rValue.UnsafePointer()

	dstSlice := (*[math.MaxInt32]byte)(dst)[:size:size]
	srcSlice := (*[math.MaxInt32]byte)(src)[:size:size]

	return copy(dstSlice, srcSlice) == int(s.itemSize)
}

func (s *storageReflect) Copy(index uint, value unsafe.Pointer) {
	dst := s.Get(index)
	size := s.itemSize

	dstSlice := (*[math.MaxInt32]byte)(dst)[:size:size]
	srcSlice := (*[math.MaxInt32]byte)(value)[:size:size]

	copy(dstSlice, srcSlice)
}

func (s *storageReflect) Shrink(cap uint) {
	if uint(s.buffer.Cap()) > cap {
		old := s.buffer
		s.buffer = reflect.New(reflect.ArrayOf(int(cap), s.typeOf)).Elem()
		s.bufferAddress = s.buffer.Addr().UnsafePointer()
		reflect.Copy(s.buffer, old)
	}
}

func (s *storageReflect) Expand(cap uint) {
	if uint(s.buffer.Cap()) < cap {
		old := s.buffer
		newSize := s.capacityIncrement * ((uint32(cap) + s.capacityIncrement) / s.capacityIncrement)
		s.buffer = reflect.New(reflect.ArrayOf(int(newSize), s.typeOf)).Elem()
		s.bufferAddress = s.buffer.Addr().UnsafePointer()
		reflect.Copy(s.buffer, old)
	}
}

func (s *storageReflect) Reset() {
	s.Shrink(0)
}

func (s storageReflect) Stats() StorageStats {
	return StorageStats{
		s.typeOf,
		uint(s.typeOf.Size()),
		uint(s.buffer.Cap()),
	}
}

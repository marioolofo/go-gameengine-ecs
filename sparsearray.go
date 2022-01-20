package ecs

// SparseArray is the interface used to map between linear indices and value indices
// Linear arrays for big structs can be a waste of memory if many entries are unused,
// so the SparseArray reduces the memory comsumption by converting between indices,
// keeping the value array contiguous and small
//
// Get returns the value Index for a given index or InvalidIndex if the index is unused
//
// Set sets the index as value
//
// Invalidate sets the index with value InvalidIndex
//
// Reset discards the mapping array and sets the SparseArray to its initial state
type SparseArray interface {
	Get(index Index) Index
	Set(index, value Index)
	Invalidate(Index)
	Reset()
}

type sparseArray struct {
	values []Index
}

type mapArray struct {
	values map[Index]Index
}

// NewSparseArray returns a new SparseArray
func NewSparseArray() SparseArray {
	arr := &sparseArray{
		make([]Index, InitialSparseArrayCapacity, InitialSparseArrayCapacity),
	}
	for i := 0; i < InitialSparseArrayCapacity; i++ {
		arr.values[i] = InvalidIndex
	}
	return arr
}

// NewSparseMapArray returns a new SparseArray with a map[index]index as internal structure
func NewSparseMapArray() SparseArray {
	return &mapArray{
		make(map[Index]Index, InitialSparseArrayCapacity),
	}
}

func (s *sparseArray) Get(index Index) Index {
	s.ensureCapacity(int(index))
	return s.values[index]
}

func (s *sparseArray) Set(index, value Index) {
	s.ensureCapacity(int(index))
	s.values[index] = value
}

func (s *sparseArray) Reset() {
	resetCapacity := InitialSparseArrayCapacity
	s.values = make([]Index, resetCapacity, resetCapacity)
	for i := 0; i < resetCapacity; i++ {
		s.values[i] = InvalidIndex
	}
}

func (s *sparseArray) Invalidate(index Index) {
	if len(s.values) > int(index) {
		s.values[index] = InvalidIndex
	}
}

func (s *sparseArray) ensureCapacity(size int) {
	length := len(s.values)
	if size >= length {
		s.values = append(s.values, make([]Index, size - length + 1) ...)
		for i := length; i <= size; i++ {
			s.values[i] = InvalidIndex
		}
	}
}

func (m *mapArray) Get(index Index) Index {
	value, exists := m.values[index]
	if !exists {
		// log.Printf("[SparseMapArray.Get] Trying to access invalid index (index: %d)\n", index)
		return InvalidIndex
	}
	return value
}

func (m *mapArray) Set(index, value Index) {
	m.values[index] = value
}

func (m *mapArray) Reset() {
	m.values = make(map[Index]Index)
}

func (m *mapArray) Invalidate(index Index) {
	m.values[index] = InvalidIndex
}

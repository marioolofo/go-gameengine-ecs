package ecs

import "testing"

func TestSparseArray(t *testing.T) {
	arr := NewSparseArray()

	t.Log("Testing SparseArray\n")
	testSparseGetSet(t, arr)
	arr.Reset()
	testSparseGetSet(t, arr)
}

func TestSparseMapArray(t *testing.T) {
	arr := NewSparseMapArray()

	t.Log("Testing SparseMapArray\n")
	testSparseGetSet(t, arr)
	arr.Reset()
	testSparseGetSet(t, arr)
}

func testSparseGetSet(t *testing.T, arr SparseArray) {
	for i := Index(0); i < 10000; i += 3 {
		arr.Set(i, i*2)
	}
	for i := Index(0); i < 10000; i++ {
		value := arr.Get(i)
		if i%3 != 0 {
			if value != InvalidIndex {
				t.Errorf("Expected InvalidIndex for index %d, found %d\n", i, value)
			}
		} else {
			if value != i*2 {
				t.Errorf("Expected %d for index %d, found %d\n", i*2, i, value)
			}
			arr.Invalidate(i)
		}
	}
	index := arr.Get(10)
	if index != InvalidIndex {
		t.Errorf("expected invalid index to return InvalidIndex, %d received\n", index)
	}
}

func BenchmarkSparseArray(b *testing.B) {
	arr := NewSparseArray()

	benchSparseGetSet(b, arr)
	arr.Reset()
	benchSparseGetSet(b, arr)
}

func BenchmarkSparseMapArray(b *testing.B) {
	arr := NewSparseMapArray()

	benchSparseGetSet(b, arr)
	arr.Reset()
	benchSparseGetSet(b, arr)
}

func benchSparseGetSet(b *testing.B, arr SparseArray) {
	for i := Index(0); i < 100000; i += 3 {
		arr.Set(i, i*2)
	}
	for i := Index(0); i < 100000; i++ {
		value := arr.Get(i)
		if i%3 != 0 {
			if value != InvalidIndex {
				b.Errorf("Expected InvalidIndex for index %d, found %d\n", i, value)
			}
		} else {
			if value != i*2 {
				b.Errorf("Expected %d for index %d, found %d\n", i*2, i, value)
			}
		}
	}
}

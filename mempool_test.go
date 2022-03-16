package ecs

import (
	"testing"
)

func TestMemPoolAllocAndRecycle(t *testing.T) {
	type timer struct {
		timeoutMS int
	}

	timerRef := timer{1000}
	mem := NewMemoryPool(0, timerRef, 0)

	newTimer, index := mem.Alloc()
	if newTimer == nil {
		t.Fatal("unable to alloc memory from MemoryPool")
	}

	getPtr := mem.Get(index + Index(100000))
	if getPtr != nil {
		t.Fatal("got valid pointer from invalid index")
	}

	getPtr = mem.Get(index)
	if newTimer != getPtr {
		t.Fatal("pointer from Alloc() and Get() are not equal")
	}

	mem.Free(index)

	newTimer, newIndex := mem.Alloc()
	if newTimer != getPtr || index != newIndex {
		t.Fatal("failed to recycle pointer")
	}

	timerPtr := (*timer)(newTimer)
	timerPtr.timeoutMS = 10

	timerPtr2 := (*timer)(mem.Get(index))
	if timerPtr2.timeoutMS != 10 {
		t.Fatal("data modification not persisted on MemoryPool")
	}
	mem.Reset()

	afterResetPtr := mem.Get(index)
	if afterResetPtr != nil {
		t.Fatal("reset failed to clear memPool structures")
	}
}

func TestMemPoolUnaligned(t *testing.T) {
	type unaligned struct {
		data [5]byte
	}
	type aligned struct {
		data [4]byte
	}

	mem := NewMemoryPool(0, aligned{}, 17)

	for i := 0; i < 5; i++ {
		ptr, _ := mem.Alloc()
		if uintptr(ptr)&1 != 0 {
			t.Errorf("alloced unaligned data: %v\n", ptr)
		}
	}
	mem.Reset()

	mem = NewMemoryPool(0, unaligned{}, 32)
	for i := 0; i < 7; i++ {
		ptr, _ := mem.Alloc()
		if uintptr(ptr)&31 != 0 {
			t.Error("alloced data with invalid alignment")
		}
	}

	mem.Reset()
}

func TestMemPoolZeroGetSet(t *testing.T) {
	type Vec struct{ x, y, z float32 }
	vec := Vec{}
	sys := NewMemoryPool(0, vec, 0)

	iterations := 100000

	// Alloc and set values directly from pointer
	for i := 0; i < iterations; i++ {
		ptrAlloc, index := sys.Alloc()
		ptr := sys.Get(index)

		if ptrAlloc != ptr {
			t.Error("ptr from Alloc() and Get() are not equal")
		}

		v := (*Vec)(ptr)
		v.x = float32(i)
		v.y = float32(i * 2)
		v.z = float32(i * 3)
	}

	// check if values are correct and update with Set()
	for i := 0; i < iterations; i++ {
		ptr := (*Vec)(sys.Get(Index(i)))

		diff := ptr.x + ptr.y + ptr.z - float32(i*6)
		if diff < -0.00001 || diff > 0.00001 {
			t.Errorf("ptr value differ from expected (%d %d %d) received (%f %f %f)\n", i, i*2, i*3, ptr.x, ptr.y, ptr.z)
		}

		vec.x = float32(i * 4)
		vec.y = float32(i * 5)
		vec.z = float32(i * 6)
		sys.Set(Index(i), &vec)
	}

	// check if values are correct and Zero() them
	for i := 0; i < iterations; i++ {
		ptr := sys.Get(Index(i))
		vecPtr := (*Vec)(ptr)

		diff := vecPtr.x + vecPtr.y + vecPtr.z - float32(i*15)
		if diff < -0.00001 || diff > 0.00001 {
			t.Errorf("vecPtr differ from expected after Set(%d %d %d) received (%f %f %f)\n", i*4, i*5, i*6, vecPtr.x, vecPtr.y, vecPtr.z)
		}

		sys.Zero(Index(i))
		if vecPtr.x != 0 || vecPtr.y != 0 || vecPtr.z != 0 {
			t.Errorf("expected zero values, found %f %f %f\n", vecPtr.x, vecPtr.y, vecPtr.z)
		}
	}

	sys.Reset()
}

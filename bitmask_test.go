package ecs

import (
	"math/rand"
	"testing"
)

func TestBitmaskNew(t *testing.T) {
	mask1 := NewMask(0, 2, 5, 7, 9)
	mask1Expected := Mask(1 | (1 << 2) | (1 << 5) | (1 << 7) | (1 << 9))

	mask2 := NewMask(2, 3, 9, 13)
	mask2Expected := Mask((1 << 2) | (1 << 3) | (1 << 9) | (1 << 13))

	if mask1 != mask1Expected {
		t.Error("Expected bitmask ", mask1Expected, ", NewMask() return is ", mask1)
	}
	if mask2 != mask2Expected {
		t.Error("Expected bitmask ", mask1Expected, ", NewMask() return is ", mask1)
	}
}

func TestBitmaskOutOfRange(t *testing.T) {
	invalidMask := NewMask(MaskTotalBits + 10)
	invalidMaskCompare := Mask(0)

	if invalidMask != invalidMaskCompare {
		t.Error("invalid mask is not equal zeroed mask", invalidMask)
	}
}

func TestBitmaskGetSetReset(t *testing.T) {
	bits := []ID{0, 1, 5, 9, 10, 30, 45, 63}
	mask := NewMask(bits...)

	for _, bit := range bits {
		if !mask.Get(bit) {
			t.Error("bit don't have true value at index ", bit)
		}
		mask.Set(bit, false)
		if mask.Get(bit) {
			t.Error("bit not set correctly at index ", bit)
		}
	}

	mask = NewMask(bits...)
	mask.Reset()

	if mask != Mask(0) {
		t.Error("mask was not reset correctly")
	}
}

func TestBitmaskContains(t *testing.T) {
	mask := NewMask(1, 2, 3, 9, 10, 15)
	valid := NewMask(1, 3, 10)
	invalid := NewMask(5, 9, 10, 32)

	if !mask.Contains(valid) {
		t.Error("valid submask don't returned true")
	}
	if mask.Contains(invalid) {
		t.Error("invalid submask don't returned false")
	}
}

func TestBitmaskSearch(t *testing.T) {
	for i := uint(0); i < MaskTotalBits; i++ {
		mask := NewMask(ID(i))
		offset := mask.NextBitSet(0)
		if offset != i {
			t.Fatalf("count bits set resulted in invalid offset: bit %d, mask (%0x) offset %d\n", i, mask, offset)
		}
	}

	testIndices := []ID{0, 2, 4, 8, 9, 20, 45}

	testBits := NewMask(testIndices...)
	index := 0
	nextBit := testBits.NextBitSet(0)
	for nextBit != MaskTotalBits {
		if nextBit != uint(testIndices[index]) {
			t.Fatalf("nextBit expecting %d, received %d\n", testIndices[index], nextBit)
		}
		index++
		nextBit = testBits.NextBitSet(nextBit + 1)
	}
}

func TestBitmaskSearchAndCount(t *testing.T) {
	randomBits := make([]uint, 10)

	randomBits = append(randomBits, 0)
	for i := 0; i < 10; i++ {
		randomBits = append(randomBits, uint(rand.Uint64()))
	}

	for _, bits := range randomBits {
		mask := Mask(bits)
		bitsSet := mask.TotalBitsSet()
		firstBitOffset := mask.NextBitSet(0)

		count := uint(0)
		for i := uint(0); i < MaskTotalBits; i++ {
			if bits&(1<<i) != 0 {
				count++
			}
		}
		if uint(bitsSet) != count {
			t.Errorf("bits set error (%0x): %d != %d\n", mask, count, bitsSet)
		}

		firstBit := uint(0)
		if mask == 0 {
			firstBit = MaskTotalBits
		} else {
			for i := mask; i != 0 && i&1 == 0; i >>= 1 {
				firstBit++
			}
		}
		if firstBit != firstBitOffset {
			t.Errorf("bit offset error (%0x): %d != %d\n", mask, firstBit, firstBitOffset)
		}
	}
}

package ecs

import (
	"math/rand"
	"testing"
)

func TestBitmaskNew(t *testing.T) {
	mask1 := MakeMask(0, 2, 5, 7, 9)
	mask1Expected := Mask{(1 | (1 << 2) | (1 << 5) | (1 << 7) | (1 << 9))}

	mask2 := MakeMask(2, 3, 9, 13)
	mask2Expected := Mask{((1 << 2) | (1 << 3) | (1 << 9) | (1 << 13))}

	if mask1 != mask1Expected {
		t.Error("Expected bitmask ", mask1Expected, ", NewMask() return is ", mask1)
	}
	if mask2 != mask2Expected {
		t.Error("Expected bitmask ", mask1Expected, ", NewMask() return is ", mask1)
	}
}

func TestBitmaskOutOfRange(t *testing.T) {
	invalidMask := MakeMask(uint64(MaskTotalBits + 10))
	invalidMaskCompare := Mask{}

	if invalidMask != invalidMaskCompare {
		t.Error("invalid mask is not equal zeroed mask", invalidMask)
	}
	if invalidMask.IsSet(uint64(MaskTotalBits + 10)) {
		t.Error("expected invalid range bit to return false", invalidMask)
	}
}

func TestBitmaskGetSetReset(t *testing.T) {
	bits := []uint64{0, 1, 5, 9, 10, 30, 45, 63, 75, 93, 117, 153, 177, 213, 241}
	mask := MakeMask(bits...)

	for _, bit := range bits {
		if !mask.IsSet(bit) {
			t.Error("bit don't have true value at index ", bit)
		}
		mask.Clear(bit)
		if mask.IsSet(bit) {
			t.Error("bit not cleared correctly at index ", bit)
		}
	}

	mask = MakeMask(bits...)
	mask.Reset()

	if (mask != Mask{}) {
		t.Error("mask was not reset correctly")
	}
}

func TestBitmaskContains(t *testing.T) {
	mask := MakeMask(1, 2, 3, 9, 10, 15)
	valid := MakeMask(1, 3, 10)
	invalid := MakeMask(5, 9, 10, 32)

	if !mask.Contains(valid) {
		t.Error("valid submask don't returned true")
	}
	if mask.Contains(invalid) {
		t.Error("invalid submask don't returned false")
	}
}

func TestBitmaskSearch(t *testing.T) {
	for i := uint64(0); i < uint64(MaskTotalBits); i++ {
		mask := MakeMask(i)
		offset := mask.NextBitSet(0)
		if offset != uint(i) {
			t.Fatalf("count bits set resulted in invalid offset: bit %d, mask (%0x) offset %d\n", i, mask, offset)
		}
	}

	testIndices := []uint64{0, 2, 4, 8, 9, 20, 45}

	testBits := MakeMask(testIndices...)
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
	randomBits := make([]uint64, 0, 20)

	for attempts := 0; attempts < 1; attempts++ {
		bitCount := uint(rand.Int63n(30))
		for i := uint(0); i < bitCount; i++ {
			randomBits = append(randomBits, uint64(rand.Int63n(256)))
		}
		mask := MakeMask(randomBits...)

		bitsSet := mask.TotalBitsSet()

		count := uint(0)
		for i := uint(0); i < MaskTotalBits; i++ {
			if mask[i>>6]&(1<<(i&63)) != 0 {
				count++
			}
		}
		if bitsSet != count {
			t.Errorf("bits set error (%0x): %d != %d\n", mask, count, bitsSet)
		}

		firstBitOffset := mask.NextBitSet(0)
		firstBit := uint(0)
		if mask.IsEmpty() {
			firstBit = MaskTotalBits
		} else {
			for _, word := range mask {
				if word == 0 {
					firstBit += 64
				} else {
					for i := word; i != 0 && i&1 == 0; i >>= 1 {
						firstBit++
					}
					break
				}
			}
		}
		if firstBit != firstBitOffset {
			t.Errorf("bit offset error (%0b): %d != %d\n", mask, firstBit, firstBitOffset)
		}

		randomBits = randomBits[:0]
	}
}

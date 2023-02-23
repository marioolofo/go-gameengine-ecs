package ecs

import "math/bits"

// MaskTotalBits is the size of Mask in bits
const MaskTotalBits = MaxComponentCount

// Mask defines an array of bits with fixed size of MaxTotalBits
type Mask [MaskTotalBits / 64]uint64

var nibbleToBitsSet = [16]uint{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

// NewMask creates a new bitmask from a list of bits
// If any bit is bigger or equal MaskTotalBits, it'll not be added to the mask
func MakeMask(bits ...uint64) Mask {
	mask := Mask{}
	for _, bit := range bits {
		mask.Set(bit)
	}
	return mask
}

func (m *Mask) Set(bit uint64) {
	if bit < uint64(MaskTotalBits) {
		m[bit>>6] |= (1 << (bit & 63))
	}
}

func (m *Mask) Clear(bit uint64) {
	if bit < uint64(MaskTotalBits) {
		m[bit>>6] &= ^(1 << (bit & 63))
	}
}

func (m Mask) IsSet(bit uint64) bool {
	if bit >= uint64(MaskTotalBits) {
		return false
	}
	return m[bit>>6]&(1<<(bit&63)) != 0
}

func (m Mask) IsEmpty() bool {
	acc := m[0]
	for i := uint64(1); i < uint64(MaskTotalBits/64); i++ {
		acc |= m[i]
	}
	return acc == 0
}

func (m *Mask) Reset() {
	*m = Mask{}
}

func (m Mask) And(mask Mask) Mask {
	newMask := m
	for i, v := range mask {
		newMask[i] &= v
	}
	return newMask
}

func (m Mask) Contains(sub Mask) bool {
	for i, v := range sub {
		if m[i]&v != v {
			return false
		}
	}
	return true
}

// TotalBitsSet returns how many bits are set in this mask
func (m Mask) TotalBitsSet() uint {
	acc := 0
	for _, v := range m {
		acc += bits.OnesCount64(v)
	}
	return uint(acc)

	// 	count := uint(0)
	//
	// 	for _, e := range m {
	// 		for e != 0 {
	// 			count += nibbleToBitsSet[e&0xf]
	// 			e >>= 4
	// 		}
	// 	}
	// 	return count
}

// NextBitSet returns the index of the next bit set in range [startingFromBit, MaskTotalBits]
// If no bit set is found within this range, the return is MaskTotalBits
// The offset at startingFromBit is checked to, so remember to use the last index found + 1 to find the next bit set
func (m Mask) NextBitSet(startingFromBit uint) uint {
	count := startingFromBit & 63
	word := startingFromBit >> 6

	e := m[word]
	e >>= count
	if e == 0 {
		if word < uint(len(m)-1) {
			return m.NextBitSet((word + 1) << 6)
		}
		return uint(MaskTotalBits)
	}
	if e&1 != 0 {
		return (word << 6) + count
	}

	count += 1

	if e&0xffffffff == 0 {
		e >>= 32
		count += 32
	}
	if e&0xffff == 0 {
		e >>= 16
		count += 16
	}
	if e&0xff == 0 {
		e >>= 8
		count += 8
	}
	if e&0xf == 0 {
		e >>= 4
		count += 4
	}
	if e&0x3 == 0 {
		e >>= 2
		count += 2
	}

	count -= uint(e & 1)

	return (word << 6) + count
}

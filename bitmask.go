package ecs

// Mask is the format of the bitmask
type Mask uint64

// Mask is the size of Mask in bits
const MaskTotalBits = 64

var nibbleToBitsSet = [16]uint{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

// NewMask creates a new bitmask from a list of IDs
// If any ID is bigger or equal MaskTotalBits, it'll not be added to the mask
func NewMask(ids ...ID) Mask {
	var mask Mask
	for _, id := range ids {
		mask.Set(id, true)
	}
	return mask
}

// Get reports if bit index defined by ID is true or false
// The return will be always false for bit >= MaskTotalBits
func (e Mask) Get(bit ID) bool {
	mask := Mask(1 << bit)
	return e&mask == mask
}

// Set sets the state of bit index to true or false
// This function has no effect for bit >= MaskTotalBits
func (e *Mask) Set(bit ID, value bool) {
	if value {
		*e |= Mask(1 << bit)
	} else {
		*e &= Mask(^(1 << bit))
	}
}

// Reset change the state of all bits to false
func (e *Mask) Reset() {
	*e = 0
}

// Contains reports if other mask is a subset of this mask
func (e Mask) Contains(other Mask) bool {
	return e&other == other
}

// TotalBitsSet returns how many bits are set in this mask
func (e Mask) TotalBitsSet() uint {
	var count uint

	for e != 0 {
		count += nibbleToBitsSet[e&0xf]
		e >>= 4
	}
	return count
}

// NextBitSet returns the index of the next bit set in range [startingFromBit, MaskTotalBits]
// If no bit set is found within this range, the return is MaskTotalBits
// The offset at startingFromBit is checked to, so remember to use the last index found + 1 to find the next bit set
func (e Mask) NextBitSet(startingFromBit uint) uint {
	count := startingFromBit

	e >>= count
	if e == 0 {
		return MaskTotalBits
	}
	if e&1 != 0 {
		return count
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

	return count
}

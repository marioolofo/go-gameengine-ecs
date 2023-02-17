package ecs

import "math/bits"

// HashEntityIDArray hashes the ID part of the EntityID array
// Uses murmur3 to hash the values
func HashEntityIDArray(values []EntityID, seed uint32) uint32 {
	hash := seed

	c1 := uint32(0xcc9e2d51)
	c2 := uint32(0x1b873593)

	for _, val := range values {
		lo := uint32(val)

		lo *= c1
		lo = bits.RotateLeft32(lo, 15)
		lo *= c2

		hash ^= lo
		hash = bits.RotateLeft32(hash, 13)
		hash = hash*5 + 0xe6546b64
	}

	hash ^= uint32(len(values))

	hash ^= hash >> 16
	hash *= 0x85ebca6b
	hash ^= hash >> 13
	hash *= 0xc2b2ae35
	hash ^= hash >> 16

	return hash
}

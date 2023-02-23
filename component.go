package ecs

// ComponentID is the type for the component identifier
type ComponentID = uint

// MakeComponentMask returns a Mask set with the bits set for the bits ComponentID list
func MakeComponentMask(bits ...ComponentID) Mask {
	mask := Mask{}
	for _, bit := range bits {
		mask.Set(uint64(bit))
	}
	return mask
}

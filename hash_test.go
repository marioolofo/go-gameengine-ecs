package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	testCases := []struct {
		desc string
		data []EntityID
		seed uint32
		hash uint32
	}{
		{
			desc: "Zeros",
			data: []EntityID{},
			hash: uint32(0),
		},
		{
			desc: "1 2 3 4",
			data: []EntityID{1, 2, 3, 4},
			hash: uint32(0x84afa2b2),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := HashEntityIDArray(tC.data, tC.seed)
			assert.Equal(t, result, tC.hash, "want %x, got %x", tC.hash, result)
		})
	}
}

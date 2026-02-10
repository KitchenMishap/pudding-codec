package scribenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"math/bits"
)

// ScribeFixedBits is a leaf node that codecs a raw n bit number (n fixed).
type FixedBits struct {
	n types.TBitCount
}

// Check that implements
var _ IScribeNode = (*FixedBits)(nil)

func NewFixedBits(n types.TBitCount) *FixedBits {
	result := FixedBits{}
	result.n = n
	return &result
}

func (fb *FixedBits) Encode(sequence []types.TSymbol,
	writer bitstream.IBitWriter) (refused bool, err error) {
	// This is because FixedBits does not encode a count
	if len(sequence) != 1 {
		panic("FixedBits can only do one at a time")
	}

	input := sequence[0]
	if bits.Len64(input) > int(fb.n) {
		return true, nil
	}

	err = writer.WriteBits(input, int(fb.n))
	if err != nil {
		return false, err
	}

	return false, nil
}

func (fb *FixedBits) Decode(reader bitstream.IBitReader) ([]types.TSymbol, error) {
	result := make([]types.TSymbol, 1)
	var err error
	result[0], err = reader.ReadBits(int(fb.n))
	return result, err
}

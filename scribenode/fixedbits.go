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

func (fb *FixedBits) Encode(symbol types.TSymbol,
	writer bitstream.IBitWriter) (refused bool, err error) {

	if bits.Len64(symbol) > int(fb.n) {
		return true, nil
	}

	err = writer.WriteBits(symbol, int(fb.n))
	if err != nil {
		return false, err
	}

	return false, nil
}

func (fb *FixedBits) Decode(reader bitstream.IBitReader) (types.TSymbol, error) {
	result, err := reader.ReadBits(int(fb.n))
	return result, err
}

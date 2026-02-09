package enginenode

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// A leaf node that codecs a raw 64 bit number
type LeafRaw struct {
	LeafNoMetaBase
}

// Check that implements
var _ IEngineNode = (*LeafRaw)(nil)

func NewLeafRaw() *LeafRaw {
	result := LeafRaw{}
	return &result
}

func (lr *LeafRaw) BidBits(sequenceProfile []alphabets.AlphabetProfile) types.TBitCount {
	if len(sequenceProfile) != 1 {
		panic("LeafRaw can only cope with one")
	}
	bitCount := types.TBitCount(0)
	for _, symbol := range sequenceProfile[0] {
		bitCount += 64 * symbol.Count
	}
	return bitCount
}

func (lr *LeafRaw) Encode(sequence []types.TSymbol,
	writer bitstream.IBitWriter) (didntKnowHow bool, err error) {
	if len(sequence) != 1 {
		panic("LeafRaw can only cope with one")
	}
	for _, v := range sequence {
		err := writer.WriteBits(v, 64)
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

func (lr *LeafRaw) Decode(reader bitstream.IBitReader) ([]types.TSymbol, error) {
	result := make([]types.TSymbol, 1)
	var err error
	result[0], err = reader.ReadBits(64)
	return result, err
}

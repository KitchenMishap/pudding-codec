package enginenode

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
)

// A handy leaf node that refuses to do anything
type LeafRefuse struct {
	LeafNoMetaBase
}

// Check that implements
var _ IEngineNode = (*LeafRefuse)(nil)

func NewLeafRefuse() *LeafRefuse {
	result := LeafRefuse{}
	return &result
}

func (lr *LeafRefuse) BidBits(_ []alphabets.AlphabetProfile) types.TBitCount {
	// This is how we say "No Bid"
	return math.MaxUint64
}

func (lr *LeafRefuse) Encode(_ []types.TSymbol,
	_ bitstream.IBitWriter) (didntKnowHow bool, err error) {
	return true, nil
}

func (lr *LeafRefuse) Decode(_ bitstream.IBitReader) ([]types.TSymbol, error) {
	return []types.TSymbol{}, nil
}

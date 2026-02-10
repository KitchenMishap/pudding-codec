package scribenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// A handy leaf node that refuses to do anything
type LeafRefuse struct {
}

// Check that implements
var _ IScribeNode = (*LeafRefuse)(nil)

func NewLeafRefuse() *LeafRefuse {
	result := LeafRefuse{}
	return &result
}

func (lr *LeafRefuse) Encode(_ []types.TSymbol,
	_ bitstream.IBitWriter) (refuse bool, err error) {
	return true, nil
}

func (lr *LeafRefuse) Decode(_ bitstream.IBitReader) ([]types.TSymbol, error) {
	return []types.TSymbol{}, nil
}

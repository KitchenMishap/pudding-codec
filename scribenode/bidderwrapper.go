package scribenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// BidderWrapper is a quick and easy way to make an IBidderScribe out of an IScribeNode
type BidderWrapper struct {
	scribe IScribeNode
}

// Check that implements
var _ IBidderScribe = (*BidderWrapper)(nil)

func WrapScribeAsBidderScribe(scribe IScribeNode) IBidderScribe {
	result := BidderWrapper{}
	result.scribe = scribe
	return &result
}

func (b *BidderWrapper) BidBits(symbol types.TSymbol) (bitCount types.TBitCount, refuse bool, err error) {
	counter := bitstream.NewBitCounter()
	refuse, err = b.scribe.Encode(symbol, counter)
	if err != nil {
		return 0, false, err
	}
	if refuse {
		return 0, true, nil
	}
	return counter.CountBits(), false, nil
}

func (b *BidderWrapper) Encode(symbol types.TSymbol, writer bitstream.IBitWriter) (bool, error) {
	return b.scribe.Encode(symbol, writer)
}
func (b *BidderWrapper) Decode(reader bitstream.IBitReader) ([]types.TSymbol, error) {
	return b.scribe.Decode(reader)
}

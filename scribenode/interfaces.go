package scribenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// IScribeNode is a specialized engine node which is not permitted
// to learn, and writes no metadata. It is therefore good for writing
// the metadata from engine nodes at a predictable cost
type IScribeNode interface {
	Encode(sequence []types.TSymbol, writer bitstream.IBitWriter) (refused bool, err error)
	Decode(reader bitstream.IBitReader) ([]types.TSymbol, error)
}

// IBidderNode is a node which can tell you how short it will encode a message.
// (A "quick and easy" way to knock up a bidder node is to wrap
// an an IScribeNode in an BidderWrapper)
type IBidderNode interface {
	BidBits(sequence []types.TSymbol) (bitcount types.TBitCount, refuse bool, err error)
}

type IBidderScribe interface {
	IScribeNode
	IBidderNode
}

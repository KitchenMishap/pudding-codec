package enginenode

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

type IEngineNode interface {
	BidBits(sequenceProfile []alphabets.AlphabetProfile) types.TBitCount
	Train(sequenceProfile []alphabets.AlphabetProfile)
	Choose()
	Encode(sequence []types.TSymbol, writer bitstream.IBitWriter) (didntKnowHow bool, err error)
	Decode(reader bitstream.IBitReader) ([]types.TSymbol, error)
	EncodeMeta(writer bitstream.IBitWriter) error
	DecodeMeta(reader bitstream.IBitReader) error
}

// Embed this for a leaf with no metadata, so you don't have to implement the
// empty fuctions
type LeafNoMetaBase struct {
	LeafBase
}

func (LeafNoMetaBase) EncodeMeta(_ bitstream.IBitWriter) error { return nil }
func (LeafNoMetaBase) DecodeMeta(_ bitstream.IBitReader) error { return nil }
func (LeafNoMetaBase) Train(_ []alphabets.AlphabetProfile)     {}

// Embed this for a leaf node (because it won't be making choices)
type LeafBase struct {
}

func (LeafBase) Choose() {}

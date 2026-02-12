package traineenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/types"
)

type TraineeWrapper struct {
	bidder scribenode.IBidderScribe
}

func WrapScribeAsTrainee(scribe scribenode.IScribeNode) ITraineeNode {
	return &TraineeWrapper{scribenode.WrapScribeAsBidderScribe(scribe)}
}

// Check that implements
var _ ITraineeNode = (*TraineeWrapper)(nil)

// A trainee wrapper doesn't do anything trainee-wise
func (tw *TraineeWrapper) Observe(_ [][]types.TSymbol) error             { return nil }
func (tw *TraineeWrapper) Improve() error                                { return nil }
func (tw *TraineeWrapper) EncodeMyMetaData(_ bitstream.IBitWriter) error { return nil }
func (tw *TraineeWrapper) DecodeMyMetaData(_ bitstream.IBitReader) error { return nil }
func (tw *TraineeWrapper) Encode(symbol types.TSymbol,
	writer bitstream.IBitWriter) (bool, error) {
	return tw.bidder.Encode(symbol, writer)
}
func (tw *TraineeWrapper) Decode(reader bitstream.IBitReader) (types.TSymbol, error) {
	return tw.bidder.Decode(reader)
}
func (tw *TraineeWrapper) BidBits(symbol types.TSymbol) (types.TBitCount, bool, error) {
	return tw.bidder.BidBits(symbol)
}

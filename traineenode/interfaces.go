package traineenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/types"
)

// ITraineeNode accepts calls related to training and
// reading/writing its own metadata
type ITraineeNode interface {
	scribenode.IBidderScribe
	// Observe a message (sequence) for the purpose of future improvement
	Observe(sequence []types.TSymbol) error
	// Improve myself based on what I have observed
	Improve() error
	EncodeMyMetaData(writer bitstream.IBitWriter) error
	DecodeMyMetaData(bitstream bitstream.IBitReader) error
}

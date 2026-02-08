package codecs

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
	"strings"
)

type CodecSelector struct {
	CodecSelectorBase
}

func NewCodecSelector(paramsCodec ICodecClass) *CodecSelector {
	result := CodecSelector{}
	result.paramsCodec = paramsCodec
	return &result
}

// CodecSelectorBase is an embeddable class that adorns your own ICodecClass with the
// ability to select from a number of encodings, according to which is cheapest bitwise,
// or a number of decodings according to some selector bits in the stream

// childCodecNames[0] is the one that encodes the selector
// childCodecNames[1..n] are codecs for each possible selection

type CodecSelectorBase struct {
	CodecCompositeBase // A selector is always a composite
	paramsCodec        ICodecClass
}

func (csb *CodecSelectorBase) Encode(data types.TData,
	writer bitstream.IBitWriter) (didntKnowHow bool, err error) {
	// First we taste the possibilities to choose the selector
	// Note that a selection of [0] corresponds to childCodecNames[1] (because [0] is the selector)
	selection, err := csb.TasteSelections(data, csb.paramsCodec)
	if err != nil {
		return false, err
	}

	// The first codec is always for encoding the selector
	selectorCodecName := csb.childCodecNames[0]
	selectorCodec, ok := csb.childCodecs[selectorCodecName]
	if !ok {
		panic("there is no selector codec (it's supposed to be at 0)")
	}
	didntKnowHow, err = selectorCodec.Encode(selection, writer)
	if err != nil {
		return false, err
	}
	if didntKnowHow {
		panic("selector codec didn't know how to encode selection")
	}

	// Offset of 1 because we don't want to "select the selector"
	selectedCodecName := csb.childCodecNames[selection+1]
	selectedCodec, ok := csb.childCodecs[selectedCodecName]
	if !ok {
		panic("selected a non-existent codec")
	}

	didntKnowHow, err = selectedCodec.Encode(data, writer)
	if err != nil {
		return false, err
	}
	if didntKnowHow {
		panic("selected codec didn't know how to encode data")
	}

	return false, nil
}

func (csb *CodecSelectorBase) Decode(reader bitstream.IBitReader) (types.TData, error) {
	// The first codec is always for decoding the selector
	selectorCodecName := csb.childCodecNames[0]
	selectorCodec, ok := csb.childCodecs[selectorCodecName]
	if !ok {
		panic("there is no selector codec (it's supposed to be at 0)")
	}

	// Decode the choice made by the selector
	codecSelection, err := selectorCodec.Decode(reader)
	if err != nil {
		return 0, err
	}

	// Find the codec chosen by the selector
	// Offset of 1 so we don't "select the selector"
	chosenCodecName := csb.childCodecNames[codecSelection+1]
	chosenCodec, ok := csb.childCodecs[chosenCodecName]
	if !ok {
		panic("the chosen codec doesn't exist")
	}

	data, err := chosenCodec.Decode(reader)
	if err != nil {
		return 0, err
	}

	return data, nil
}

// Returns a selection code into the possible codecs
// If it returns zero, it really means selectorCodecNames[1] (because [0] is the selector
// codec itself for encoding/decoding the choice)
func (csb *CodecSelectorBase) TasteSelections(data types.TData,
	paramsCodec ICodecClass) (types.TIndex, error) {
	// Back up the selector codec (the one that encodes the choice)
	selectorCodecName := csb.childCodecNames[0]
	selectorCodec := csb.childCodecs[selectorCodecName]
	selectorBackup, selectorBackupSize, err := backupCodec(paramsCodec, selectorCodec)
	if err != nil {
		return 0, err
	}

	// Iterate through other child codecs
	winner := types.TIndex(math.MaxInt64)
	winnerBitCost := uint64(math.MaxUint64)
	for i, name := range csb.childCodecNames[1:] {
		selection := types.TIndex(i) // 0 "means" childCodecNames[1] (because 0 is the selector codec)
		codec := csb.childCodecs[name]

		// Back up its current state
		codecBackup, codecBackupSize, err := backupCodec(paramsCodec, codec)

		// A place to encode the data
		sb := strings.Builder{}
		bitWriter := bitstream.NewBitWriter(&sb)
		bitCounter := bitstream.NewBitCounterPassThrough(bitWriter)

		// Don't forget that there's a cost associated with SELECTING this codec
		// (subtly, this cost might even vary depending on which selection is chosen)
		didntKnowHow, err := selectorCodec.Encode(selection, bitCounter)
		if err != nil {
			return 0, err
		}
		if didntKnowHow {
			return 0, errors.New("selector codec didn't know how to select")
		}

		didntKnowHow, err = codec.Encode(data, bitCounter)
		if err != nil {
			return 0, err
		}
		if didntKnowHow {
			continue
		} // This can't be the winner

		// Count the cost. This also involves testing the metadata size
		cost := bitCounter.CountBits()
		_, newCodecMetadataSize, err := backupCodec(paramsCodec, codec)
		if err != nil {
			return 0, err
		}
		metadataCost := newCodecMetadataSize - codecBackupSize
		cost += metadataCost
		_, newSelectorMetadataSize, err := backupCodec(paramsCodec, selectorCodec)
		if err != nil {
			return 0, err
		}
		selectorMetadataCost := newSelectorMetadataSize - selectorBackupSize
		cost += selectorMetadataCost

		// Put the codec's metadata back the way we found it
		err = restoreCodec(paramsCodec, codec, codecBackup)
		if err != nil {
			return 0, err
		}

		// Put the selector codec's metadata back the way we found it
		err = restoreCodec(paramsCodec, selectorCodec, selectorBackup)
		if err != nil {
			return 0, err
		}

		// Are we winning?
		if cost < winnerBitCost {
			winner = selection
			winnerBitCost = cost
		}
	} // for codecs

	// The winning selection
	return winner, nil
}

func backupCodec(paramsCodec ICodecClass,
	codec ICodecClass) (backup string, bitCount types.TCount, err error) {
	sb := strings.Builder{}
	bitWriter := bitstream.NewBitWriter(&sb)
	bitCounter := bitstream.NewBitCounterPassThrough(bitWriter)
	err = codec.WriteParams(paramsCodec, bitCounter)
	if err != nil {
		return "", types.TCount(0), err
	}
	backup = sb.String()
	bitCount = bitCounter.CountBits()
	return backup, bitCount, nil
}

func restoreCodec(paramsCodec ICodecClass, codec ICodecClass, backup string) error {
	sr := strings.NewReader(backup)
	bitReader := bitstream.NewBitReader(sr)
	return codec.ReadParams(paramsCodec, bitReader)
}

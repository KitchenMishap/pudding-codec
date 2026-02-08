package compression

import (
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/codecs"
	"github.com/KitchenMishap/pudding-codec/types"
	"io"
)

type Engine struct {
	paramsCodec codecs.ICodecClass // For metadata
	dataCodec   codecs.ICodecClass // For data
}

// Check that implements
var _ IEngine = (*Engine)(nil)

func NewEngine(paramsCodec codecs.ICodecClass, rootCodec codecs.ICodecClass) *Engine {
	result := Engine{}
	result.paramsCodec = paramsCodec
	result.dataCodec = rootCodec
	return &result
}

func (eng *Engine) Encode(data []types.TData, writer io.Writer) error {
	bitWriter := bitstream.NewBitWriter(writer)
	bitCounter := bitstream.NewBitCounterPassThrough(bitWriter)

	// Initially we'd train the engine; we haven't implemented that yet

	// First we have to write the metadata (params) using the paramsCodec
	// (of course the params codec can't itself have metadata! we'd have a chicken and egg)
	err := eng.dataCodec.WriteParams(eng.paramsCodec, bitCounter)
	if err != nil {
		return err
	}

	// Secondly we write the actual data (which weirdly includes the length)
	length := len(data)
	didntKnowHow, err := eng.dataCodec.Encode(types.TData(length), bitCounter)
	if err != nil {
		return err
	}
	if didntKnowHow {
		return errors.New("Engine.Encode(), length, encoder did not know how")
	}

	for _, value := range data {
		didntKnowHow, err = eng.dataCodec.Encode(value, bitCounter)
		if err != nil {
			return err
		}
		if didntKnowHow {
			return errors.New("Engine.Encode(), data, encoder did not know how")
		}
	}
	err = bitWriter.FlushBits()
	if err != nil {
		return err
	}

	startGB := bitstream.FormatBits(8 * 8 * uint64(len(data)))
	endGB := bitstream.FormatBits(bitCounter.CountBits())
	percent := 100 * float64(bitCounter.CountBits()) / float64(8*8*len(data))
	fmt.Printf("Compressed %s into %s, %.2f%%\n", startGB, endGB, percent)

	return nil
}

func (eng *Engine) Decode(reader io.Reader) (data []types.TData, err error) {
	bitReader := bitstream.NewBitReader(reader)

	// First we have to read the metadata
	err = eng.dataCodec.ReadParams(eng.paramsCodec, bitReader)
	if err != nil {
		return nil, err
	}

	// Secondly we read the data itself, which weirdly includes the length
	length, err := eng.dataCodec.Decode(bitReader)
	if err != nil {
		return nil, err
	}

	result := make([]types.TData, length)
	for i := range length {
		result[i], err = eng.dataCodec.Decode(bitReader)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

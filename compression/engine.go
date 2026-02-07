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
	composite codecs.ICodecClass
}

// Check that implements
var _ IEngine = (*Engine)(nil)

func NewEngine() *Engine {
	result := Engine{}
	result.composite = codecs.NewCompositeExample()
	return &result
}

func (eng *Engine) Encode(data []types.TData, writer io.Writer) (err error) {
	bitWriter := bitstream.NewBitWriter(writer)
	bitCounter := bitstream.NewBitCounterPassThrough(bitWriter)

	length := len(data)
	didntKnowHow, err := eng.composite.Encode(types.TData(length), bitCounter)
	if err != nil {
		return err
	}
	if didntKnowHow {
		return errors.New("Engine.Encode(), length, encoder did not know how")
	}

	for _, value := range data {
		didntKnowHow, err = eng.composite.Encode(value, bitCounter)
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

	length, err := eng.composite.Decode(bitReader)
	if err != nil {
		return nil, err
	}
	result := make([]types.TData, length)
	for i := range length {
		result[i], err = eng.composite.Decode(bitReader)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

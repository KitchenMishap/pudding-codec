package compression

import (
	"errors"
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

	length := len(data)
	didntKnowHow, err := eng.composite.Encode(types.TData(length), bitWriter)
	if err != nil {
		return err
	}
	if didntKnowHow {
		return errors.New("Engine.Encode(), length, encoder did not know how")
	}

	for _, value := range data {
		didntKnowHow, err = eng.composite.Encode(value, bitWriter)
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

package compression

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/codecs"
	"github.com/KitchenMishap/pudding-codec/types"
	"io"
)

type Engine struct {
	theCodecs codecs.ICodecCollection
	streamer  bitstream.IBitStream
}

// Check that implements
var _ IEngine = (*Engine)(nil)

func NewEngine() *Engine {
	result := Engine{}
	result.theCodecs = codecs.NewCodecSlice()
	result.theCodecs.AddCodec(codecs.NewCodecRaw64())
	result.streamer = bitstream.NewStringSlice()
	return &result
}

func (eng *Engine) Encode(data []types.TData, writer io.Writer) error {
	bitWriter := bitstream.NewBitWriter(writer)

	length := len(data)
	err := eng.theCodecs.GetCodec(0).Encode(types.TData(length), bitWriter)
	if err != nil {
		return err
	}

	for _, value := range data {
		err = eng.theCodecs.GetCodec(0).Encode(value, bitWriter)
		if err != nil {
			return err
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

	length, err := eng.theCodecs.GetCodec(0).Decode(bitReader)
	if err != nil {
		return nil, err
	}
	result := make([]types.TData, length)
	for i := range length {
		result[i], err = eng.theCodecs.GetCodec(0).Decode(bitReader)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

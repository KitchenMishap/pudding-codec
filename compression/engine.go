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
	length := len(data)
	bitCode := eng.theCodecs.GetCodec(0).Encode(types.TData(length))
	err := eng.streamer.PushBack(bitCode)
	if err != nil {
		return err
	}
	for _, value := range data {
		bitCode = eng.theCodecs.GetCodec(0).Encode(value)
		err = eng.streamer.PushBack(bitCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (eng *Engine) Decode(reader io.Reader) (data []types.TData, err error) {
	bitCode, err := eng.streamer.PopFront(64)
	if err != nil {
		return nil, err
	}
	length := eng.theCodecs.GetCodec(0).Decode(bitCode)
	result := make([]types.TData, length)
	for i := range length {
		bitCode, err := eng.streamer.PopFront(64)
		if err != nil {
			return nil, err
		}
		result[i] = eng.theCodecs.GetCodec(0).Decode(bitCode)
	}
	return result, nil
}

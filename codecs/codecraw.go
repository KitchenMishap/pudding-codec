package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitcode"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

type CodecRaw64 struct {
}

// Check that implements
var _ ICodecClass = (*CodecRaw64)(nil)

func NewCodecRaw64() *CodecRaw64 {
	return &CodecRaw64{}
}

func (codec *CodecRaw64) ReadParams(initializedCodecs []ICodecClass,
	stream bitstream.IBitStream) {
	// Nothing to read, CodecRaw64 has no parameters
}

func (codec *CodecRaw64) Encode(data types.TData) bitcode.IBitCode {
	return bitcode.NewBitCode64(uint64(data), 64)
}

func (codec *CodecRaw64) Decode(bitCode bitcode.IBitCode) types.TData {
	return bitCode.Bits()
}

func (codec *CodecRaw64) WriteParams(initializedCodecs []ICodecClass,
	stream bitstream.IBitStream) {
	// Nothing to write, CodecRaw64 has no parameters
}

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

func (codec *CodecRaw64) Encode(data types.TData, bitWriter bitstream.IBitWriter) error {
	bitCode := bitcode.NewBitCode64(uint64(data), 64)
	return bitWriter.WriteBits(bitCode.Bits(), bitCode.Length())
}

func (codec *CodecRaw64) Decode(bitReader bitstream.IBitReader) (types.TData, error) {
	bits, err := bitReader.ReadBits(64)
	if err != nil {
		return 0, err
	}
	return bits, nil
}

func (codec *CodecRaw64) WriteParams(initializedCodecs []ICodecClass,
	stream bitstream.IBitStream) {
	// Nothing to write, CodecRaw64 has no parameters
}

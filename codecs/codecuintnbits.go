package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

type CodecUintNBits struct {
	N int // The parameter, how many bits to read/write
}

// Check that implements
var _ ICodecClass = (*CodecUintNBits)(nil)

func NewCodecUintNBits(n int) *CodecUintNBits {
	return &CodecUintNBits{n}
}

func (nb *CodecUintNBits) WriteParams(paramsCodec ICodecClass, stream bitstream.IBitWriter) error {
	didntKnowHow, err := paramsCodec.Encode(types.TData(nb.N), stream)
	if didntKnowHow {
		panic("params codec didn't know how to encode n")
	}
	return err
}

func (nb *CodecUintNBits) ReadParams(paramsCodec ICodecClass, stream bitstream.IBitReader) error {
	val, err := paramsCodec.Decode(stream)
	nb.N = int(val)
	return err
}

func (nb *CodecUintNBits) Encode(data types.TData, writer bitstream.IBitWriter) (didntKnowHow bool, err error) {
	if data >= (1 << uint(nb.N)) {
		return true, nil
	}
	return false, writer.WriteBits(data, nb.N)
}

func (nb *CodecUintNBits) Decode(reader bitstream.IBitReader) (data types.TData, err error) {
	data, err = reader.ReadBits(nb.N)
	return data, err
}

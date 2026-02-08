package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// Encodes each symbol as N bits, where N is held in metadata

type SymbolCodecNBits struct {
	N int // The parameter, how many bits to read/write
}

// Check that implements
var _ ICodecClass = (*SymbolCodecNBits)(nil)

func NewSymbolCodecNBits(n int) *SymbolCodecNBits {
	return &SymbolCodecNBits{n}
}

func (nb *SymbolCodecNBits) WriteParams(paramsCodec ICodecClass, stream bitstream.IBitWriter) error {
	didntKnowHow, err := paramsCodec.Encode(types.TData(nb.N), stream)
	if didntKnowHow {
		panic("params codec didn't know how to encode n")
	}
	return err
}

func (nb *SymbolCodecNBits) ReadParams(paramsCodec ICodecClass, stream bitstream.IBitReader) error {
	val, err := paramsCodec.Decode(stream)
	nb.N = int(val)
	return err
}

func (nb *SymbolCodecNBits) Encode(data types.TSymbol, writer bitstream.IBitWriter) (didntKnowHow bool, err error) {
	if data >= (1 << uint(nb.N)) {
		return true, nil
	}
	return false, writer.WriteBits(data, nb.N)
}

func (nb *SymbolCodecNBits) Decode(reader bitstream.IBitReader) (data types.TSymbol, err error) {
	data, err = reader.ReadBits(nb.N)
	return data, err
}

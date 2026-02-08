package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

type ICodecClass interface {
	// Write my parameters (metadata) to stream using any of the supplied codecs
	WriteParams(paramsCodec ICodecClass, stream bitstream.IBitWriter) error
	// Read my own parameters (metadata) from stream
	// I may need to the use the prior supplied codecs to decode my params
	ReadParams(paramsCodec ICodecClass, stream bitstream.IBitReader) error
	// Encode a value into a bitcode using my parameters. This is my speciality
	Encode(data types.TData, writer bitstream.IBitWriter) (didntKnowHow bool, err error)
	// Decode a value from a bitcode
	Decode(reader bitstream.IBitReader) (types.TData, error)
}

type ICodecSelectorClass interface {
	ICodecClass
	SelectionCount() types.TCount
}

type INamedCodecCollection interface {
	AddCodec(codec ICodecClass, name string) error
	GetCodec(name string) (ICodecClass, error)
}

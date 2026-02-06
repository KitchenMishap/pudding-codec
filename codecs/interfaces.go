package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitcode"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

type ICodecClass interface {
	// Read my own parameters (metadata) from stream
	// I may need to the use the prior supplied codecs to decode my params
	ReadParams(initializedCodecs []ICodecClass, stream bitstream.IBitStream)
	// Encode a value into a bitcode using my parameters. This is my speciality
	Encode(val types.TData) bitcode.IBitCode
	// Decode a value from a bitcode
	Decode(bitCode bitcode.IBitCode) types.TData
	// Write my parameters (metadata) to stream using any of the supplied codecs
	WriteParams(initializedCodecs []ICodecClass, stream bitstream.IBitStream)
}

type ICodecCollection interface {
	Count() int
	AddCodec(codec ICodecClass)
	GetCodec(idx int) ICodecClass
}

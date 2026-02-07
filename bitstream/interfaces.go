package bitstream

import "github.com/KitchenMishap/pudding-codec/bitcode"

type IBitStream interface {
	PushBack(code bitcode.IBitCode) error
	PopFront(bitcount int) (bitcode.IBitCode, error)
}

type IBitWriter interface {
	WriteBits(bits uint64, bitCount int) error
	FlushBits() error
}

type IBitReader interface {
	ReadBits(bitCount int) (uint64, error)
}

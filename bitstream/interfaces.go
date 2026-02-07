package bitstream

import "github.com/KitchenMishap/pudding-codec/bitcode"

type IBitStream interface {
	PushBack(code bitcode.IBitCode) error
	PopFront() (bitcode.IBitCode, error)
}

package bitstream

import "github.com/KitchenMishap/pudding-codec/bitcode"

type IBitStream interface {
	PushBack(code bitcode.IBitCode) error
	Pop() (bitcode.IBitCode, error)
}

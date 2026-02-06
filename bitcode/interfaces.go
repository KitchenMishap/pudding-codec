package bitcode

import "io"

type IBitCode interface {
	Length() int
	Bits() uint64
	ReadBytes(io.Reader) error
	WriteBytes(io.Writer) error
}

package bitstream

type IBitWriter interface {
	WriteBits(bits uint64, bitCount int) error
	FlushBits() error
}

type IBitReader interface {
	ReadBits(bitCount int) (uint64, error)
}

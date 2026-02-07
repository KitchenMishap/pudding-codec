package bitstream

import (
	"encoding/binary"
	"errors"
	"io"
)

type BitReader struct {
	byteReader io.Reader
}

// Check that implements
var _ IBitReader = (*BitReader)(nil)

type BitWriter struct {
	byteWriter io.Writer
}

// Check that implements
var _ IBitWriter = (*BitWriter)(nil)

func NewBitReader(byteReader io.Reader) *BitReader {
	return &BitReader{byteReader: byteReader}
}
func NewBitWriter(byteWriter io.Writer) *BitWriter {
	return &BitWriter{byteWriter: byteWriter}
}

func (br *BitReader) ReadBits(bitCount int) (uint64, error) {
	if bitCount%8 != 0 {
		panic("not implemented")
	}
	byteCount := bitCount / 8
	bytes := [8]byte{}
	n, err := br.byteReader.Read(bytes[:byteCount])
	if err != nil {
		return 0, err
	}
	if n != byteCount {
		return 0, errors.New("didn't read all bytes")
	}
	return binary.LittleEndian.Uint64(bytes[:]), nil
}

func (bw *BitWriter) WriteBits(bits uint64, bitCount int) error {
	if bitCount%8 != 0 {
		panic("not implemented")
	}
	byteCount := bitCount / 8
	bytes := [8]byte{}
	binary.LittleEndian.PutUint64(bytes[:], bits)
	n, err := bw.byteWriter.Write(bytes[:byteCount])
	if err != nil {
		return err
	}
	if n != byteCount {
		return errors.New("didn't write all bytes")
	}
	return nil
}

func (bw *BitWriter) FlushBits() error {
	return nil
}

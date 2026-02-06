package bitcode

import (
	"encoding/binary"
	"errors"
	"github.com/KitchenMishap/pudding-codec/utils"
	"io"
)

type BitCode64 struct {
	bits   uint64
	length int
}

// Check that implements
var _ IBitCode = (*BitCode64)(nil)

func NewBitCode64(bits uint64, length int) *BitCode64 {
	result := BitCode64{}
	result.bits = bits
	result.length = length
	return &result
}

func (bc *BitCode64) Length() int {
	return bc.length
}

func (bc *BitCode64) Bits() uint64 {
	return bc.bits
}

func (bc *BitCode64) WriteBytes(writer io.Writer) error {
	nineBytes := [9]byte{}
	length := bc.Length()
	if length > 64 {
		panic("can only cope with max 64 bits")
	}

	nineBytes[0] = byte(length)

	binary.LittleEndian.PutUint64(nineBytes[1:], bc.bits)

	byteCount := utils.BucketCount(int64(length), 8)
	written, err := writer.Write(nineBytes[:1+byteCount])
	if err != nil {
		return err
	}
	if written != 1+int(byteCount) {
		return errors.New("did not write all bytes")
	}
	return nil
}

func (bc *BitCode64) ReadBytes(reader io.Reader) error {
	nineBytes := [9]byte{}
	read, err := reader.Read(nineBytes[0:1])
	if err != nil {
		return err
	}
	if read != 1 {
		return errors.New("did not read 1 byte")
	}

	bc.length = int(nineBytes[0])
	byteCount := int(utils.BucketCount(int64(bc.Length()), 8))

	read, err = reader.Read(nineBytes[1 : byteCount+1])
	if err != nil {
		return err
	}
	if read != byteCount {
		return errors.New("did not read all bytes")
	}
	bc.bits = binary.LittleEndian.Uint64(nineBytes[1 : 1+8])
	return nil
}

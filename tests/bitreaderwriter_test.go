package tests

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"os"
	"testing"
)

func Test_BitReaderWriter(t *testing.T) {
	fName := "../TestingFiles/bits.bin"
	byteWriter, err := os.Create(fName)
	if err != nil {
		t.Error(err)
	}

	bitWriter := bitstream.NewBitWriter(byteWriter)

	num8bit := uint64(255)
	num16bit := uint64(32768)
	err = bitWriter.WriteBits(num8bit, 8)
	if err != nil {
		t.Error(err)
	}
	err = bitWriter.WriteBits(num16bit, 16)
	if err != nil {
		t.Error(err)
	}
	err = bitWriter.FlushBits()
	if err != nil {
		t.Error(err)
	}
	err = byteWriter.Close()
	if err != nil {
		t.Error(err)
	}

	byteReader, err := os.Open(fName)
	if err != nil {
		t.Error(err)
	}
	bitReader := bitstream.NewBitReader(byteReader)

	num8bitBack, err := bitReader.ReadBits(8)
	if err != nil {
		t.Error(err)
	}
	num16bitBack, err := bitReader.ReadBits(16)
	if err != nil {
		t.Error(err)
	}
	err = byteReader.Close()
	if err != nil {
		t.Error(err)
	}

	if num8bitBack != num8bit {
		t.Error("8 bit number doesn't match")
	}
	if num16bitBack != num16bit {
		t.Error("8 bit number doesn't match")
	}
}

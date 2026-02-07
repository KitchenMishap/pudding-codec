package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"os"
	"testing"
)

func TestCodecSlice(t *testing.T) {
	// Prepare
	fName := "../TestingFiles/testCodecSlice.bin"
	paramsCodec := NewCodecRaw64()
	file, err := os.Create(fName)
	if err != nil {
		t.Error(err)
	}
	bitWriter := bitstream.NewBitWriter(file)

	// Write
	num1 := types.TData(123456789)
	dontKnowHow, err := paramsCodec.Encode(num1, bitWriter)
	if err != nil {
		t.Error(err)
	}
	if dontKnowHow {
		t.Error("Don't Know How")
	}

	num2 := types.TData(987654321)
	dontKnowHow, err = paramsCodec.Encode(num2, bitWriter)
	if err != nil {
		t.Error(err)
	}
	if dontKnowHow {
		t.Error("Don't Know How")
	}
	err = bitWriter.FlushBits()
	if err != nil {
		t.Error(err)
	}
	err = file.Close()
	if err != nil {
		t.Error(err)
	}

	// Read
	file, err = os.Open(fName)
	bitReader := bitstream.NewBitReader(file)
	if err != nil {
		t.Error(err)
	}
	newNum1, err := paramsCodec.Decode(bitReader)
	if err != nil {
		t.Error(err)
	}
	newNum2, err := paramsCodec.Decode(bitReader)

	if newNum1 != num1 {
		t.Errorf("newNum1 != num1: %d != %d", newNum1, num1)
	}
	if newNum2 != num2 {
		t.Errorf("newNum2 != num2: %d != %d", newNum2, num2)
	}
}

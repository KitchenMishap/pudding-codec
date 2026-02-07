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
	codeclist := NewCodecSlice()
	codeclist.AddCodec(NewCodecRaw64())
	rawIndex := 0
	file, err := os.Create(fName)
	if err != nil {
		t.Error(err)
	}
	bitWriter := bitstream.NewBitWriter(file)

	// Write
	num1 := types.TData(123456789)
	err = codeclist.GetCodec(rawIndex).Encode(num1, bitWriter)
	if err != nil {
		t.Error(err)
	}
	num2 := types.TData(987654321)
	err = codeclist.GetCodec(rawIndex).Encode(num2, bitWriter)
	if err != nil {
		t.Error(err)
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
	newNum1, err := codeclist.GetCodec(rawIndex).Decode(bitReader)
	if err != nil {
		t.Error(err)
	}
	newNum2, err := codeclist.GetCodec(rawIndex).Decode(bitReader)

	if newNum1 != num1 {
		t.Errorf("newNum1 != num1: %d != %d", newNum1, num1)
	}
	if newNum2 != num2 {
		t.Errorf("newNum2 != num2: %d != %d", newNum2, num2)
	}
}

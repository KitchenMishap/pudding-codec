package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"testing"
)

func TestFifo(t *testing.T) {
	// Prepare
	codeclist := NewCodecSlice()
	codeclist.AddCodec(NewCodecRaw64())
	rawIndex := 0
	stream := bitstream.StringSlice{}

	// Write
	num1 := types.TData(123456789)
	bitCode := codeclist.GetCodec(rawIndex).Encode(num1)
	err := stream.PushBack(bitCode)
	if err != nil {
		t.Error(err)
	}
	num2 := types.TData(987654321)
	bitCode2 := codeclist.GetCodec(rawIndex).Encode(num2)
	err = stream.PushBack(bitCode2)
	if err != nil {
		t.Error(err)
	}

	// Read
	newBitCode1, err := stream.PopFront()
	if err != nil {
		t.Error(err)
	}
	newNum1 := codeclist.GetCodec(rawIndex).Decode(newBitCode1)
	newBitCode2, err := stream.PopFront()
	if err != nil {
		t.Error(err)
	}
	newNum2 := codeclist.GetCodec(rawIndex).Decode(newBitCode2)

	if newNum1 != num1 {
		t.Errorf("newNum1 != num1: %d != %d", newNum1, num1)
	}
	if newNum2 != num2 {
		t.Errorf("newNum2 != num2: %d != %d", newNum2, num2)
	}
}

package codecs

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"os"
	"testing"
)

func Test_CodecLut64(t *testing.T) {
	codecLut := NewCodecLut64()

	num1 := types.TData(12345)
	num2 := types.TData(54321)
	codecLut.teachValue(num1)
	codecLut.teachValue(num2)

	// Codec for writing params
	codecParams := NewCodecRaw64()

	// Stream
	fName := "../TestingFiles/codecLut64.bin"
	file, err := os.Create(fName)
	if err != nil {
		t.Error(err)
	}
	bitWriter := bitstream.NewBitWriter(file)

	// Write params
	err = codecLut.WriteParams(codecParams, bitWriter)
	if err != nil {
		t.Error(err)
	}

	// Write data
	dontKnowHow, err := codecLut.Encode(num1, bitWriter)
	if err != nil {
		t.Error(err)
	}
	if dontKnowHow {
		t.Error("should know how")
	}
	dontKnowHow, err = codecLut.Encode(num1, bitWriter)
	if err != nil {
		t.Error(err)
	}
	if dontKnowHow {
		t.Error("should know how")
	}
	dontKnowHow, err = codecLut.Encode(num2, bitWriter)
	if err != nil {
		t.Error(err)
	}
	if dontKnowHow {
		t.Error("should know how")
	}

	num3 := types.TData(987654321)
	dontKnowHow, err = codecLut.Encode(num3, bitWriter)
	if err != nil {
		t.Error(err)
	}
	if !dontKnowHow {
		t.Error("shouldn't know how")
	}

	err = bitWriter.FlushBits()
	if err != nil {
		t.Error(err)
	}
	err = file.Close()
	if err != nil {
		t.Error(err)
	}

	// Done writing, now read

	file2, err := os.Open(fName)
	if err != nil {
		t.Error(err)
	}
	bitReader := bitstream.NewBitReader(file2)

	// Read params
	codecLut2 := NewCodecLut64()
	err = codecLut2.ReadParams(codecParams, bitReader)
	if err != nil {
		t.Error(err)
	}

	// Read Values
	numBack1, err := codecLut2.Decode(bitReader)
	if err != nil {
		t.Error(err)
	}
	numBack2, err := codecLut2.Decode(bitReader)
	if err != nil {
		t.Error(err)
	}
	numBack3, err := codecLut2.Decode(bitReader)
	if err != nil {
		t.Error(err)
	}

	// Check they match
	if numBack1 != num1 {
		t.Error("numBack1 != num1")
	}
	if numBack2 != num1 {
		t.Error("numBack2 != num1")
	}
	if numBack3 != num2 {
		t.Error("numBack3 != num2")
	}

	err = file2.Close()
	if err != nil {
		t.Error(err)
	}
}

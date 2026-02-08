package tests

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/codecs"
	"github.com/KitchenMishap/pudding-codec/types"
	"os"
	"testing"
)

/*
func Test_SelectorExample(t *testing.T) {
	inputData := []types.TData{1, 65537, 3}

	// Prepare a params codec (this could be the root of a tree)
	paramsCodec := codecs.NewCodecRaw64()

	// Prepare a data codec (this could be the root of a tree)
	// This codec has a codec for the selection, and some others to choose from
	dataCodec := codecs.NewSelectorExample(paramsCodec)
	selectionCodec := codecs.NewCodecRaw64()            // This is the codec that encodes the "which" number
	err := dataCodec.AddCodec(selectionCodec, "Select") // This MUST be the first codec of the selector
	if err != nil {
		t.Error("failed to add selection codec")
	}
	rawCodec := codecs.NewCodecRaw64() // One of the selections is "raw"
	err = dataCodec.AddCodec(rawCodec, "Raw")
	if err != nil {
		t.Error("failed to add raw codec")
	}
	lutCodec := codecs.NewCodecLut64() // One of the selections is "lut"
	err = dataCodec.AddCodec(lutCodec, "Lut")
	if err != nil {
		t.Error("failed to add lut codec")
	}

	// Make the engine out of the above two codec trees
	engine := compression.NewEngine(paramsCodec, dataCodec)

	fName := "../TestingFiles/selectorTest.bin"
	file, err := os.Create(fName)
	if err != nil {
		t.Error(err)
	}
	err = engine.Encode(inputData, file)
	if err != nil {
		t.Error(err)
	}
	err = file.Close()
	if err != nil {
		t.Error(err)
	}

	file, err = os.Open(fName)
	if err != nil {
		t.Error(err)
	}
	outputData, err := engine.Decode(file)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(inputData, outputData) {
		t.Errorf("encode error, expected %v, got %v", inputData, outputData)
	}
}*/

func TestCodecComposite(t *testing.T) {
	// Prepare
	fName := "../TestingFiles/testCodecSlice.bin"
	paramsCodec := codecs.NewCodecRaw64()
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

func Test_CodecLut64(t *testing.T) {
	codecLut, err := codecs.NewCodecLut64(8)
	if err != nil {
		t.Error(err)
	}

	num1 := types.TData(12345)
	num2 := types.TData(54321)
	codecLut.TeachValue(num1)
	codecLut.TeachValue(num2)

	// Codec for writing params
	codecParams := codecs.NewCodecRaw64()

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
	codecLut2, err := codecs.NewCodecLut64(8)
	if err != nil {
		t.Error(err)
	}
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

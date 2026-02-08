package tests

import (
	"github.com/KitchenMishap/pudding-codec/codecs"
	"github.com/KitchenMishap/pudding-codec/compression"
	"github.com/KitchenMishap/pudding-codec/types"
	"os"
	"reflect"
	"testing"
)

func Test_SelectorRawLut(t *testing.T) {
	inputData := []types.TData{
		123456789, 123456789, 123456789, 123456789, 123456789, 123456789, 123456789,
		987654321, 987654321, 987654321, 987654321, 987654321, 987654321, 987654321}

	// Prepare a params codec (this could be the root of a tree)
	paramsCodec := codecs.NewCodecRaw64()
	// Prepare a data codec (this could be the root of a tree)
	dataCodec, err := codecs.NewSelectorRawLutCombo(paramsCodec)
	if err != nil {
		t.Error(err)
	}
	// Make the engine out of the above two codec trees
	engine := compression.NewEngine(paramsCodec, dataCodec)

	fName := "../TestingFiles/test.bin"
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
}

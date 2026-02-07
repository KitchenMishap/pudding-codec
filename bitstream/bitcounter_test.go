package bitstream

import (
	"github.com/KitchenMishap/pudding-codec/compression"
	"github.com/KitchenMishap/pudding-codec/types"
	"os"
	"reflect"
	"testing"
)

func Test_Engine(t *testing.T) {
	inputData := []types.TData{1, 2, 3}

	engine := compression.NewEngine()

	fName := "../TestingFiles/engineTest.bin"
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

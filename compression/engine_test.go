package compression

import (
	"github.com/KitchenMishap/pudding-codec/types"
	"reflect"
	"testing"
)

func Test_Engine(t *testing.T) {
	inputData := []types.TData{1, 2, 3}

	engine := NewEngine()

	err := engine.Encode(inputData, nil)
	if err != nil {
		t.Error(err)
	}
	outputData, err := engine.Decode(nil)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(inputData, outputData) {
		t.Errorf("encode error, expected %v, got %v", inputData, outputData)
	}
}

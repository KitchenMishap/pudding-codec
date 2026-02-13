package tests

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/compositeroots"
	engine2 "github.com/KitchenMishap/pudding-codec/engine"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
	"os"
	"reflect"
	"testing"
)

func Test_DemographerEncodeDecode(t *testing.T) {
	data := []types.TData{1, 2, 3, 0, types.TData(math.MaxUint64), 123456, 1, 2, 3, 123456}

	metaDataRootDemo := compositeroots.NewRawScribe()
	dataRootDemo := compositeroots.NewDemographerTrainee()
	engineDemo := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootDemo),
		dataRootDemo)

	err := engineDemo.DataNode.Observe(data)
	if err != nil {
		t.Fatal(err)
	}
	err = engineDemo.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("../TestingFiles/test.bin")
	if err != nil {
		t.Fatal(err)
	}
	bw := bitstream.NewBitWriter(file)
	refused, err := engineDemo.Encode(data, bw)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("trained engine refused")
	}
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	file2, err := os.Open("../TestingFiles/test.bin")
	if err != nil {
		t.Fatal(err)
	}
	br := bitstream.NewBitReader(file2)
	metaDataRootDemo2 := compositeroots.NewRawScribe()
	dataRootDemo2 := compositeroots.NewDemographerTrainee()
	engineDemo2 := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootDemo2),
		dataRootDemo2)

	dataOut, err := engineDemo2.Decode(br)
	if err != nil {
		t.Fatal(err)
	}
	err = file2.Close()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, dataOut) {
		t.Errorf("encode error, expected %v, got %v", data, dataOut)
	}
}

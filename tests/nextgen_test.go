package tests

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/compositeroots"
	engine2 "github.com/KitchenMishap/pudding-codec/engine"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/traineenode"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
	"os"
	"reflect"
	"testing"
)

func Test_RawVersusChoiceScribe(t *testing.T) {
	data := []types.TData{1, 2, 3, math.MaxUint64}
	dataBits := types.TBitCount(64 * len(data))

	// Raw
	metaDataRootRaw := compositeroots.NewRawScribe()
	dataRootRaw := compositeroots.NewRawScribe()
	engineRaw := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRaw),
		traineenode.WrapScribeAsTrainee(dataRootRaw))

	rawBits, refused, err := engineRaw.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("Engine refused")
	}
	if rawBits != dataBits {
		t.Fatal("wrong bits bid")
	}

	// Choice
	metaDataRootChoice := compositeroots.NewChoiceScribe()
	dataRootChoice := compositeroots.NewChoiceScribe()
	engineChoice := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootChoice),
		traineenode.WrapScribeAsTrainee(dataRootChoice))

	choiceBits, refused, err := engineChoice.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("Engine refused")
	}
	if choiceBits >= dataBits {
		t.Fatal("bits didn't improve")
	}

}

func Test_LeafRefuse(t *testing.T) {
	data := []types.TData{1}

	metaDataRoot := scribenode.WrapScribeAsBidderScribe(scribenode.NewLeafRefuse())
	dataRoot := traineenode.WrapScribeAsTrainee(scribenode.NewLeafRefuse())
	engine := engine2.NewNextGenEngine(metaDataRoot, dataRoot)

	_, refused, err := engine.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if !refused {
		t.Fatal("Engine should have refused")
	}
}

func Test_RawVersusTraineeChoiceScribe(t *testing.T) {
	data := []types.TData{1, 2, 3, math.MaxUint64, 123456, 1, 2, 3, 123456}
	dataBits := types.TBitCount(64 * len(data))

	// Raw
	metaDataRootRaw := compositeroots.NewRawScribe()
	dataRootRaw := compositeroots.NewRawScribe()
	engineRaw := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRaw),
		traineenode.WrapScribeAsTrainee(dataRootRaw))

	rawBits, refused, err := engineRaw.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("Engine refused")
	}
	if rawBits != dataBits {
		t.Fatal("wrong bits bid")
	}

	// Trainee Choice
	metaDataRootChoice := compositeroots.NewChoiceScribe()
	dataRootChoice := compositeroots.NewChoiceTrainee()
	engineChoice := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootChoice),
		traineenode.WrapScribeAsTrainee(dataRootChoice))

	untrainedBits, refused, err := engineChoice.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("untrained engine refused")
	}

	err = engineChoice.DataNode.Observe(data)
	if err != nil {
		t.Fatal(err)
	}
	err = engineChoice.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	trainedBits, refused, err := engineChoice.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("trained engine refused")
	}
	if trainedBits > untrainedBits {
		t.Fatal("bits got worse")
	}
}

func Test_TraineeChoiceVersusShannonFano(t *testing.T) {
	data := []types.TData{1, 2, 3, math.MaxUint64, 123456, 1, 2, 3, 123456}

	// Trainee Choice
	metaDataRootChoice := compositeroots.NewChoiceScribe()
	dataRootChoice := compositeroots.NewChoiceTrainee()
	engineChoice := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootChoice),
		traineenode.WrapScribeAsTrainee(dataRootChoice))

	_, refused, err := engineChoice.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("untrained engine refused")
	}

	err = engineChoice.DataNode.Observe(data)
	if err != nil {
		t.Fatal(err)
	}
	err = engineChoice.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	trainedBits, refused, err := engineChoice.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("trained engine refused")
	}

	// Shannon Fano
	metaDataRootShannon := compositeroots.NewChoiceScribe()
	dataRootShannon := compositeroots.NewShannonFanoTrainee()
	engineShannon := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootShannon),
		dataRootShannon)

	_, refused, err = engineShannon.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if !refused {
		t.Fatal("untrained engine should refuse")
	}

	err = engineShannon.DataNode.Observe(data)
	if err != nil {
		t.Fatal(err)
	}
	err = engineShannon.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	trainedBitsShannon, refused, err := engineShannon.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("trained engine refused")
	}
	if trainedBitsShannon >= trainedBits {
		t.Fatal("bits didn't improve")
	}
}

func Test_ShannonEncodeDecode(t *testing.T) {
	data := []types.TData{1, 2, 3, math.MaxUint64, 123456, 1, 2, 3, 123456}

	metaDataRootShannon := compositeroots.NewChoiceScribe()
	dataRootShannon := compositeroots.NewShannonFanoTrainee()
	engineShannon := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootShannon),
		dataRootShannon)

	err := engineShannon.DataNode.Observe(data)
	if err != nil {
		t.Fatal(err)
	}
	err = engineShannon.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("../TestingFiles/test.bin")
	if err != nil {
		t.Fatal(err)
	}
	bw := bitstream.NewBitWriter(file)
	refused, err := engineShannon.Encode(data, bw)
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
	metaDataRootShannon2 := compositeroots.NewChoiceScribe()
	dataRootShannon2 := compositeroots.NewShannonFanoTrainee()
	engineShannon2 := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootShannon2),
		dataRootShannon2)

	dataOut, err := engineShannon2.Decode(br)
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

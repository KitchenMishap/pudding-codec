package tests

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/compositeroots"
	engine2 "github.com/KitchenMishap/pudding-codec/engine"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/sequences"
	"github.com/KitchenMishap/pudding-codec/traineenode"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
	"os"
	"reflect"
	"testing"
)

func Test_RawVersusRoundedDecimal(t *testing.T) {
	data := []types.TData{1, 2, 3, math.MaxUint32, 123456, 1, 2, 3, 123456}

	// Raw =========================================
	metaDataRootRaw := compositeroots.NewRawScribe()
	dataRootRaw := compositeroots.NewRawScribe()
	engineRaw := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRaw),
		traineenode.WrapScribeAsTrainee(dataRootRaw))

	rawBid, refused, err := engineRaw.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("untrained engine refused")
	}

	// RoundDecimal =======================================
	metaDataRootRound := compositeroots.NewRawScribe()
	dataRootRound := compositeroots.NewRoundDecimalTrainee()
	engineRound := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRound),
		dataRootRound)

	_, refused, err = engineRound.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if !refused {
		t.Fatal("untrained engine should refuse")
	}

	err = engineRound.DataNode.Observe(
		sequences.SingleSymbolDataToSampleOfSequences(data))
	if err != nil {
		t.Fatal(err)
	}
	err = engineRound.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	roundBid, refused, err := engineRound.BidBits(data)
	if err != nil {
		t.Fatal(err)
	}
	if refused {
		t.Fatal("trained engine refused")
	}
	if roundBid >= rawBid {
		t.Fatal("bits didn't improve")
	}
}

func Test_RoundObserveZeroTest(t *testing.T) {
	metaDataRootRound := compositeroots.NewRawScribe()
	dataRootRound := compositeroots.NewRoundDecimalTrainee()
	engineRound := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRound),
		dataRootRound)

	err := engineRound.DataNode.Observe(
		sequences.SingleSymbolDataToSampleOfSequences([]types.TData{0}))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_RoundEncodeDecode(t *testing.T) {
	data := []types.TData{1, 2, 3, 0, math.MaxUint32, 123456, 1, 2, 3, 123456}

	metaDataRootRound := compositeroots.NewRawScribe()
	dataRootRound := compositeroots.NewRoundDecimalTrainee()
	engineRound := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRound),
		dataRootRound)

	err := engineRound.DataNode.Observe(
		sequences.SingleSymbolDataToSampleOfSequences(data))
	if err != nil {
		t.Fatal(err)
	}
	err = engineRound.DataNode.Improve()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("../TestingFiles/test.bin")
	if err != nil {
		t.Fatal(err)
	}
	bw := bitstream.NewBitWriter(file)
	refused, err := engineRound.Encode(data, bw)
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
	metaDataRootRound2 := compositeroots.NewRawScribe()
	dataRootRound2 := compositeroots.NewRoundDecimalTrainee()
	engineRound2 := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootRound2),
		dataRootRound2)

	dataOut, err := engineRound2.Decode(br)
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

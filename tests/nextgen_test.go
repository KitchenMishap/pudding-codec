package tests

import (
	"github.com/KitchenMishap/pudding-codec/compositeroots"
	engine2 "github.com/KitchenMishap/pudding-codec/engine"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/traineenode"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
	"testing"
)

func Test_RawVersusChoiceScribe(t *testing.T) {
	data := []types.TData{1, 2, 3, math.MaxUint32}
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
		t.Fatal("Engine should haverefused")
	}
}

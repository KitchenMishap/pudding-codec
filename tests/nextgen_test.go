package tests

import (
	engine2 "github.com/KitchenMishap/pudding-codec/engine"
	"github.com/KitchenMishap/pudding-codec/enginenode"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
	"testing"
)

func Test_LeafRaw(t *testing.T) {
	data := []types.TData{1, 2, 3}
	dataBits := types.TBitCount(64 * 3)

	metaDataRoot := enginenode.NewLeafRaw()
	dataRoot := enginenode.NewLeafRaw()
	engine := engine2.NewNextGenEngine(metaDataRoot, dataRoot)

	bidBits := engine.BidBits(data)
	if bidBits != dataBits {
		t.Fatal("wrong bits bid")
	}
}

func Test_LeafRefuse(t *testing.T) {
	data := []types.TData{1, 2, 3}

	metaDataRoot := enginenode.NewLeafRaw()
	dataRoot := enginenode.NewLeafRefuse()
	engine := engine2.NewNextGenEngine(metaDataRoot, dataRoot)

	bidBits := engine.BidBits(data)
	if bidBits != math.MaxUint64 {
		t.Fatal("wrong bits bid")
	}
}

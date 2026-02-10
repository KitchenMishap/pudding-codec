package engine

import (
	"github.com/KitchenMishap/pudding-codec/enginenode"
	"github.com/KitchenMishap/pudding-codec/types"
)

type NextGenEngine struct {
	MetaDataNode enginenode.IMetaDataNode
	DataNode     enginenode.IEngineNode
}

func NewNextGenEngine(metaDataNode enginenode.IMetaDataNode, dataNode enginenode.IEngineNode) *NextGenEngine {
	result := NextGenEngine{metaDataNode, dataNode}
	return &result
}

func (ng *NextGenEngine) BidBits(sequence []types.TSymbol) (bitCount types.TBitCount, refused bool, err error) {
	count := types.TBitCount(0)
	for _, symbol := range sequence {
		countBits, refused, err := ng.DataNode.BidBits([]types.TSymbol{symbol})
		if err != nil {
			return 0, false, err
		}
		if refused {
			return 0, true, nil
		}
		count += countBits
	}
	return count, false, nil
}

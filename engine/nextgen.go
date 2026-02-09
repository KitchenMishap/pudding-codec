package engine

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/enginenode"
	"github.com/KitchenMishap/pudding-codec/types"
)

type NextGenEngine struct {
	MetaDataNode enginenode.IEngineNode
	DataNode     enginenode.IEngineNode
}

func NewNextGenEngine(metaDataNode enginenode.IEngineNode, dataNode enginenode.IEngineNode) *NextGenEngine {
	result := NextGenEngine{metaDataNode, dataNode}
	return &result
}

func (ng *NextGenEngine) BidBits(data alphabets.DataSet) types.TBitCount {
	sequenceProfile := make([]alphabets.AlphabetProfile, 1)
	sequenceProfile[0], _ = alphabets.AlphabetProfileFromData(data)
	return ng.DataNode.BidBits(sequenceProfile)
}

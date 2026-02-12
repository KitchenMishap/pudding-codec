package engine

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
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
		countBits, refused, err := ng.DataNode.BidBits(symbol)
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

func (ng *NextGenEngine) Encode(sequence []types.TSymbol, writer bitstream.IBitWriter) (refused bool, err error) {
	// This metadata is the result of training
	err = ng.DataNode.EncodeMyMetaData(writer)
	if err != nil {
		return false, err
	}

	// Encode the count as metadata
	refused, err = ng.MetaDataNode.Encode(types.TSymbol(len(sequence)), writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("metadata node refused to encode count")
	}

	for _, symbol := range sequence {
		refused, err := ng.DataNode.Encode(symbol, writer)
		if err != nil {
			return false, err
		}
		if refused {
			return true, nil
		}
	}
	err = writer.FlushBits()
	return false, err
}

func (ng *NextGenEngine) Decode(reader bitstream.IBitReader) (sequence []types.TSymbol, err error) {
	// This metadata is the result of training
	err = ng.DataNode.DecodeMyMetaData(reader)
	if err != nil {
		return nil, err
	}

	count, err := ng.MetaDataNode.Decode(reader)
	if err != nil {
		return nil, err
	}
	result := make([]types.TSymbol, int(count))
	for i := range count {
		symbol, err := ng.DataNode.Decode(reader)
		if err != nil {
			return nil, err
		}
		result[i] = symbol
	}
	return result, nil
}

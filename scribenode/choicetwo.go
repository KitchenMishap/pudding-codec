package scribenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

type ChoiceTwo struct {
	switchNode  IScribeNode
	optionNodes [2]IBidderScribe
}

func NewChoiceTwo(switchNode IScribeNode, optionNodes [2]IBidderScribe) *ChoiceTwo {
	return &ChoiceTwo{switchNode, optionNodes}
}

func (ct *ChoiceTwo) Encode(sequence []types.TSymbol, writer bitstream.IBitWriter) (refused bool, err error) {
	// Assess
	cost0, refuse0, err := ct.optionNodes[0].BidBits(sequence)
	if err != nil {
		return false, err
	}
	cost1, refuse1, err := ct.optionNodes[1].BidBits(sequence)
	if err != nil {
		return false, err
	}

	// Choose
	switchSymbol := types.TSymbol(0)
	if refuse0 || cost1 < cost0 {
		switchSymbol = types.TSymbol(1)
	}
	// If both choices refused, WE refuse
	if refuse0 && refuse1 {
		return true, nil
	}

	// Encode switch
	switchSequence := []types.TSymbol{switchSymbol} // Sequence of 1 symbol
	refused, err = ct.switchNode.Encode(switchSequence, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ChoiceTwo: switch refused to encode")
	}

	// Encode sequence
	refused, err = ct.optionNodes[switchSymbol].Encode(sequence, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ChoiceTwo: option node refused to encode")
	}

	return false, nil
}

func (ct *ChoiceTwo) Decode(reader bitstream.IBitReader) ([]types.TSymbol, error) {
	// Read the switch
	switchSequence, err := ct.switchNode.Decode(reader)
	if err != nil {
		return []types.TSymbol{}, err
	}

	return ct.optionNodes[switchSequence[0]].Decode(reader)
}

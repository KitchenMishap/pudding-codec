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

func (ct *ChoiceTwo) Encode(symbol types.TSymbol, writer bitstream.IBitWriter) (refused bool, err error) {
	// Assess
	cost0, refuse0, err := ct.optionNodes[0].BidBits(symbol)
	if err != nil {
		return false, err
	}
	cost1, refuse1, err := ct.optionNodes[1].BidBits(symbol)
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
	refused, err = ct.switchNode.Encode(switchSymbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ChoiceTwo: switch refused to encode")
	}

	// Encode sequence
	refused, err = ct.optionNodes[switchSymbol].Encode(symbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ChoiceTwo: option node refused to encode")
	}

	return false, nil
}

func (ct *ChoiceTwo) Decode(reader bitstream.IBitReader) (types.TSymbol, error) {
	// Read the switch
	switchSymbol, err := ct.switchNode.Decode(reader)
	if err != nil {
		return 0, err
	}
	return ct.optionNodes[switchSymbol].Decode(reader)
}

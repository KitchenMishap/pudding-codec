package compositeroots

import (
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/traineenode"
)

func NewChoiceTrainee() scribenode.IScribeNode {
	switchPositionNode := scribenode.NewFixedBits(1)
	option0 := scribenode.NewFixedBits(8)
	option1 := scribenode.NewFixedBits(64)
	optionNodes := [2]scribenode.IBidderScribe{
		scribenode.WrapScribeAsBidderScribe(option0),
		scribenode.WrapScribeAsBidderScribe(option1),
	}
	choiceNode := traineenode.NewChoiceTwo(switchPositionNode, optionNodes)
	return choiceNode
}

func NewShannonFanoTrainee() traineenode.ITraineeNode {
	switchPositionNode := scribenode.NewFixedBits(1)
	choiceNode := traineenode.NewRecursiveShannonFano(switchPositionNode)
	return choiceNode
}

func NewRoundDecimalTrainee() traineenode.ITraineeNode {
	leadingZerosNode := NewShannonFanoTrainee()
	metaDigitsNode := NewShannonFanoTrainee()
	roundNode := traineenode.NewRoundishDecimal(leadingZerosNode, metaDigitsNode)
	return roundNode
}

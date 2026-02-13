package compositeroots

import (
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/traineenode"
	"github.com/KitchenMishap/pudding-codec/types"
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

func NewRoundDecimalTrainee(rate types.TData) traineenode.ITraineeNode {
	rateMultiplierScribe := NewRawScribe()
	leadingZerosNode := NewShannonFanoTrainee()
	metaDigitsNode := NewShannonFanoTrainee()
	roundNode := traineenode.NewRoundishDecimal(rateMultiplierScribe,
		leadingZerosNode, metaDigitsNode, rate)
	return roundNode
}

func NewDemographerTrainee() traineenode.ITraineeNode {
	rateNodes := []traineenode.ITraineeNode{
		NewRoundDecimalTrainee(1),
		NewRoundDecimalTrainee(2),
		NewRoundDecimalTrainee(3),
		NewRoundDecimalTrainee(4),
		NewRoundDecimalTrainee(5),
		NewRoundDecimalTrainee(6),
		NewRoundDecimalTrainee(7),
		NewRoundDecimalTrainee(8),
		NewRoundDecimalTrainee(9),
	}
	switchNode := NewShannonFanoTrainee()
	demoNode := traineenode.NewDemographer(switchNode, rateNodes)
	return demoNode
}

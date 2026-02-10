package compositeroots

import (
	"github.com/KitchenMishap/pudding-codec/scribenode"
)

func NewRawScribe() scribenode.IScribeNode {
	return scribenode.NewFixedBits(64)
}

func NewChoiceScribe() scribenode.IScribeNode {
	switchRepresentationNode := scribenode.NewFixedBits(1)
	option0 := scribenode.NewFixedBits(8)
	option1 := scribenode.NewFixedBits(64)
	optionNodes := [2]scribenode.IBidderScribe{
		scribenode.WrapScribeAsBidderScribe(option0),
		scribenode.WrapScribeAsBidderScribe(option1),
	}
	choiceNode := scribenode.NewChoiceTwo(switchRepresentationNode, optionNodes)
	return choiceNode
}

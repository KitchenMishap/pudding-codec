package enginenode

import (
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/traineenode"
)

type IEngineNode interface {
	traineenode.ITraineeNode // So I can receive calls related to training
}

type IMetaDataNode interface {
	scribenode.IBidderScribe // So I can Encode / Decode / Bid. That's all
}

package tests

import (
	"github.com/KitchenMishap/pudding-codec/traineenode"
	"testing"
)

func Test_Eleven(t *testing.T) {
	eleven := uint64(11)
	traineenode.EatLeadingRepeatingDigit(eleven, 2, 10)
}

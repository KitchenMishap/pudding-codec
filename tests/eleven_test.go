package tests

import (
	"github.com/KitchenMishap/pudding-codec/traineenode"
	"testing"
)

func Test_Eleven(t *testing.T) {
	eleven := uint64(11)
	traineenode.EatLeadingRepeatingDigit(eleven, 2, 10)
}

func Test_1420520461(t *testing.T) {
	onefourtwo := uint64(1420520461)
	traineenode.RoundishNumberRepresentation(2 * onefourtwo)
}

func Test_12(t *testing.T) {
	onetwotwo := uint64(122)
	dig, rep, _, scrapsdigits, _ := traineenode.EatLeadingRepeatingDigit(onetwotwo, 20, 10_000_000_000_000_000_000)
	if dig != 0 {
		t.Error("should hav found zeros")
	}
	if rep != 20-3 {
		t.Error("should have found 17 zeros")
	}
	if scrapsdigits != 3 {
		t.Error("should be three digits left")
	}
	traineenode.RoundishNumberRepresentation(onetwotwo)
}

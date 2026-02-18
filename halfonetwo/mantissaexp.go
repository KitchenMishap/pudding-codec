package halfonetwo

import (
	"math/bits"
)

type MantissaExponentTallies struct {
	mantissaShift        int
	mantissaMask         uint64
	mantissaCount        uint64
	ExpTalliesByMantissa [][64]uint64
}

func NewMantissaExponentTallies(mantissaBitCount int) *MantissaExponentTallies {
	if mantissaBitCount < 1 || mantissaBitCount > 64 {
		panic("silly mantissa bitcount")
	}
	// Some useful information based on mantissaBitCount
	result := MantissaExponentTallies{}
	result.mantissaCount = 1 << mantissaBitCount
	result.mantissaShift = 64 - mantissaBitCount
	mask := result.mantissaCount - 1
	result.mantissaMask = mask << result.mantissaShift

	// Make the histograms
	result.ExpTalliesByMantissa = make([][64]uint64, result.mantissaCount)
	return &result
}

func (met *MantissaExponentTallies) Wipe() {
	for mantissa := range met.mantissaCount {
		for leadingZeros := range 64 {
			met.ExpTalliesByMantissa[mantissa][leadingZeros] = 0
		}
	}
}

func (met *MantissaExponentTallies) Populate(amounts []uint64) {
	for _, amount := range amounts {
		if amount == 0 {
			continue
		} // Not interested in zeros!
		leadingZeros := bits.LeadingZeros64(amount)
		mantissa := amount << leadingZeros
		mantissaBin := (mantissa & met.mantissaMask) >> met.mantissaShift
		met.ExpTalliesByMantissa[mantissaBin][leadingZeros] += 1
	}
}

type DominantAmount struct {
	Amount   uint64
	Strength int
}

func (met *MantissaExponentTallies) AnalyzeHalfOneTwo() []DominantAmount {
	result := make([]DominantAmount, 0)
	for mantissa := range met.mantissaCount {
		for leadingZerosForOne := 1; leadingZerosForOne <= 62; leadingZerosForOne++ {
			leadingZerosForHalf := leadingZerosForOne + 1 // The "half" has one more leading zero than the "one"
			leadingZerosForTwo := leadingZerosForOne - 1  // The "two" has one less leading zero than the "one"
			halfTally := met.ExpTalliesByMantissa[mantissa][leadingZerosForHalf]
			oneTally := met.ExpTalliesByMantissa[mantissa][leadingZerosForOne]
			twoTally := met.ExpTalliesByMantissa[mantissa][leadingZerosForTwo]
			if halfTally*oneTally*twoTally == 0 {
				continue // Not interested unless all non-zero
			}
			if halfTally > oneTally || twoTally > oneTally {
				continue // Not interested unless the "one" tally dominates
			}
			strength := halfTally // Slowest ship in a convoy
			if twoTally > halfTally {
				strength = twoTally
			}
			// We have a candidate half-one-two pattern featuring "strength" triplets of amounts

			// We have the mantissa, what is the amount?
			mantissaInTopBits := mantissa << met.mantissaShift
			originalAmount := mantissaInTopBits >> leadingZerosForOne
			result = append(result, DominantAmount{originalAmount, int(strength)})
		}
	}
	return result
}

func FilterHalfOneTwoFiveTenTwenty(halfOneTwoResults []DominantAmount) []DominantAmount {
	result := make([]DominantAmount, 0)
	// Examine every pair
	for indexA, halfOneTwoA := range halfOneTwoResults {
		// To go with "A", we only look at "B"s that are before it
		for indexB := 0; indexB < indexA; indexB++ {
			halfOneTwoB := halfOneTwoResults[indexB]

			// So we have two different entries, and we're only visiting each pair once
			amountA := float64(halfOneTwoA.Amount)
			amountB := float64(halfOneTwoB.Amount)
			ratio := amountA / amountB

			// Not quite sure of the logic for multiplying strengths here... it just "feels" right!
			combinedStrength := halfOneTwoA.Strength * halfOneTwoB.Strength

			if ratio > 0.099 && ratio < 0.101 {
				// amountB is about ten times amountA... splendid! A suitable pair
				if halfOneTwoA.Strength > halfOneTwoB.Strength {
					// amountA has stronger evidence. Store that one
					result = append(result, DominantAmount{halfOneTwoA.Amount, combinedStrength})
				} else {
					// amountB is stronger
					result = append(result, DominantAmount{halfOneTwoB.Amount, combinedStrength})
				}
			} else if ratio > 9.9 && ratio < 10.1 {
				// amountA is about ten times amountB... splendid! A suitable pair
				if halfOneTwoA.Strength > halfOneTwoB.Strength {
					// amountA has stronger evidence. Store that one
					result = append(result, DominantAmount{halfOneTwoA.Amount, combinedStrength})
				} else {
					// amountB is stronger
					result = append(result, DominantAmount{halfOneTwoB.Amount, combinedStrength})
				}
			}
		}
	}
	return result
}

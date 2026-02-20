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
			if oneTally < strength {
				strength = oneTally
			}
			if twoTally < strength {
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

// Like the above, but no "one" to insist on being strongest
func (met *MantissaExponentTallies) AnalyzeEighthQuarterHalf() []DominantAmount {
	result := make([]DominantAmount, 0)
	for mantissa := range met.mantissaCount {
		for leadingZerosForQuarter := 1; leadingZerosForQuarter <= 62; leadingZerosForQuarter++ {
			leadingZerosForEighth := leadingZerosForQuarter + 1 // The "eighth" has one more leading zero than the "quarter"
			leadingZerosForHalf := leadingZerosForQuarter - 1   // The "half" has one less leading zero than the "quarter"
			eighthTally := met.ExpTalliesByMantissa[mantissa][leadingZerosForEighth]
			quarterTally := met.ExpTalliesByMantissa[mantissa][leadingZerosForQuarter]
			halfTally := met.ExpTalliesByMantissa[mantissa][leadingZerosForHalf]
			if eighthTally*quarterTally*halfTally == 0 {
				continue // Not interested unless all non-zero
			}
			strength := eighthTally // Slowest ship in a convoy
			if quarterTally < strength {
				strength = quarterTally
			}
			if halfTally < strength {
				strength = halfTally
			}
			// We have a candidate eighth-quarter-half pattern featuring "strength" triplets of amounts

			// We have the mantissa, what is the amount?
			mantissaInTopBits := mantissa << met.mantissaShift
			originalAmount := mantissaInTopBits >> leadingZerosForQuarter
			result = append(result, DominantAmount{originalAmount, int(strength)})
		}
	}
	return result
}

// Also (unfortuneately) tends to match 100_200_250_400_500_1000
func Filter50_100_125_200_250_500(triplets50_100_200 []DominantAmount,
	triples125_250_500 []DominantAmount) []DominantAmount {
	result := make([]DominantAmount, 0)
	// Examine every pair
	for _, fifty_100_200 := range triplets50_100_200 {
		for _, onetwentyfive_250_500 := range triples125_250_500 {
			// So we have two different entries, and we're only visiting each pair once
			amountA := float64(fifty_100_200.Amount)
			amountB := float64(onetwentyfive_250_500.Amount)
			ratio := amountB / amountA

			// Not quite sure of the logic for multiplying strengths here... it just "feels" right!
			combinedStrength := fifty_100_200.Strength * onetwentyfive_250_500.Strength

			// 250 is 2.5 times 100
			if ratio > 2.49 && ratio < 2.51 {
				// splendid! A suitable pair
				result = append(result, DominantAmount{fifty_100_200.Amount, combinedStrength})
			}
		}
	}
	return result
}

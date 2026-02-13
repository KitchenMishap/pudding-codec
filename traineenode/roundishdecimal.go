package traineenode

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/types"
	"github.com/KitchenMishap/pudding-codec/utils"
)

type RoundishDecimal struct {
	rateMultiplierScribe scribenode.IScribeNode
	leadingZerosNode     ITraineeNode
	metaDigitNode        ITraineeNode

	rateMultiplier types.TData // Data will be multiplied by this before rounding is attempted

	// No stored observations (except in children nodes)
	// No stored metadata (except in children nodes)
}

// Check that implements
var _ ITraineeNode = (*RoundishDecimal)(nil)

func NewRoundishDecimal(rateMultiplierScribe scribenode.IScribeNode,
	leadingZerosNode ITraineeNode, metaDigitNode ITraineeNode,
	rateMultiplier types.TData) *RoundishDecimal {
	result := RoundishDecimal{}
	result.rateMultiplierScribe = rateMultiplierScribe
	result.leadingZerosNode = leadingZerosNode
	result.metaDigitNode = metaDigitNode
	result.rateMultiplier = rateMultiplier
	return &result
}

func (rd *RoundishDecimal) Encode(value types.TData, writer bitstream.IBitWriter) (refused bool, err error) {
	_, overflow := utils.SafeMultiply(value, rd.rateMultiplier)
	if overflow {
		return true, nil
	}
	symbolSequence := RoundishNumberRepresentation(value * rd.rateMultiplier)
	if len(symbolSequence) < 1 {
		panic("not enough symbols in sequence")
	}

	leadingZerosSymbol := symbolSequence[0]
	refused, err = rd.leadingZerosNode.Encode(leadingZerosSymbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("leadingZeros encoder refused")
	}

	for _, metaDigitSymbol := range symbolSequence[1:] {
		refused, err = rd.metaDigitNode.Encode(metaDigitSymbol, writer)
		if err != nil {
			return false, err
		}
		if refused {
			panic("metaDigit encoder refused")
		}
	}
	return false, nil
}
func (rd *RoundishDecimal) Decode(reader bitstream.IBitReader) (types.TSymbol, error) {
	// First decode the leading zeros count
	leadingZerosSymbol, err := rd.leadingZerosNode.Decode(reader)
	if err != nil {
		return 0, err
	}
	leadingZerosCount := LeadingZerosFrom(leadingZerosSymbol)

	// Start with an empty slice of meta digits
	metaDigits := make([]MetaDigit, 0, 10)
	for NeedAnotherMetaDigit(leadingZerosCount, metaDigits) {
		// Decode another meta digit
		metaDigitSymbol, err := rd.metaDigitNode.Decode(reader)
		if err != nil {
			return 0, err
		}
		metaDigits = append(metaDigits, RepeatingDigitFrom(metaDigitSymbol))
	}
	// Reconstruct the number, starting at the least significant digit
	total := types.TData(0)
	powTen := types.TData(1)
	for i := len(metaDigits) - 1; i >= 0; i-- {
		metaDigit := metaDigits[i]
		for range metaDigit.RepeatCount {
			total += types.TData(metaDigit.Digit) * powTen
			powTen *= 10
		}
	}
	return total / rd.rateMultiplier, nil
}

func (rd *RoundishDecimal) BidBits(value types.TSymbol) (bitCount types.TBitCount, refused bool, err error) {
	bitCount = types.TBitCount(0)

	_, overflow := utils.SafeMultiply(value, rd.rateMultiplier)
	if overflow {
		return 0, true, nil
	}
	symbolSequence := RoundishNumberRepresentation(value * rd.rateMultiplier)
	if len(symbolSequence) < 1 {
		panic("not enough symbols in sequence")
	}

	leadingZerosSymbol := symbolSequence[0]
	subCount, refused, err := rd.leadingZerosNode.BidBits(leadingZerosSymbol)
	if err != nil {
		return 0, false, err
	}
	if refused {
		return 0, true, nil
	}
	bitCount += subCount

	for _, metaDigitSymbol := range symbolSequence[1:] {
		subCount, refused, err = rd.metaDigitNode.BidBits(metaDigitSymbol)
		if err != nil {
			return 0, false, err
		}
		if refused {
			return 0, true, nil
		}
		bitCount += subCount
	}
	return bitCount, false, nil
}

func (rd *RoundishDecimal) Observe(samples []types.TSymbol) error {
	for _, sample := range samples {
		_, overflow := utils.SafeMultiply(sample, rd.rateMultiplier)
		if overflow {
			continue
		} // Don't observe overflows. They'll refuse to encode anyway

		symbolSequence := RoundishNumberRepresentation(sample * rd.rateMultiplier)
		if len(symbolSequence) < 1 {
			panic("too few symbols")
		}

		leadingZerosSymbol := symbolSequence[0]
		err := rd.leadingZerosNode.Observe([]types.TSymbol{leadingZerosSymbol})
		if err != nil {
			return err
		}

		for _, metaDigitSymbol := range symbolSequence[1:] {
			err = rd.metaDigitNode.Observe([]types.TSymbol{metaDigitSymbol})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rd *RoundishDecimal) Improve() error {
	// Improve the leading zeros node
	err := rd.leadingZerosNode.Improve()
	if err != nil {
		return err
	}
	// And the metaDigit node
	err = rd.metaDigitNode.Improve()
	return err
}

func (rd *RoundishDecimal) EncodeMyMetaData(writer bitstream.IBitWriter) error {
	// My metadata
	refused, err := rd.rateMultiplierScribe.Encode(rd.rateMultiplier, writer)
	if err != nil {
		return err
	}
	if refused {
		return errors.New("RoundishDecimal: rate multiplier metadata node refused to encode")
	}

	// My children's metadata
	err = rd.leadingZerosNode.EncodeMyMetaData(writer)
	if err != nil {
		return err
	}
	err = rd.metaDigitNode.EncodeMyMetaData(writer)
	if err != nil {
		return err
	}
	return nil
}
func (rd *RoundishDecimal) DecodeMyMetaData(reader bitstream.IBitReader) error {
	// My metadata
	var err error
	rd.rateMultiplier, err = rd.rateMultiplierScribe.Decode(reader)
	if err != nil {
		return err
	}

	// My children's metadata
	err = rd.leadingZerosNode.DecodeMyMetaData(reader)
	if err != nil {
		return err
	}
	err = rd.metaDigitNode.DecodeMyMetaData(reader)
	if err != nil {
		return err
	}
	return nil
}

// The protocol dictates 20 decimal digits, including leading zeros
// (enough to fill a uint64) but the symbols each represent different numbers of digits
const digitsNeeded = 20

func RoundishNumberRepresentation(number types.TData) []types.TSymbol {
	symbolsResult := make([]types.TSymbol, 0, 10)

	// The first power of 10 we are interested is the biggest power of 10 representable in a uint64
	powerOf10 := types.TData(10_000_000_000_000_000_000) // 1 with 19 zeros
	digits := digitsNeeded

	doingLeadingZeros := true
	scraps := number
	scrapsDigitCount := digits
	for scraps > 0 || doingLeadingZeros {
		var repeatingDigit, repeatCount int
		repeatingDigit, repeatCount, scraps, scrapsDigitCount, powerOf10 = EatLeadingRepeatingDigit(scraps, scrapsDigitCount, powerOf10)
		if doingLeadingZeros {
			if repeatingDigit == 0 {
				// The first digit was (one or more) zeros
				symbolsResult = append(symbolsResult, LeadingZerosSymbol(repeatCount)) // Some leading zeros
				doingLeadingZeros = false
			} else {
				// The first digit was (one or more) non-zeros
				// We MUSt still say there were NO leading zeros
				symbolsResult = append(symbolsResult, LeadingZerosSymbol(0))
				symbolsResult = append(symbolsResult, Symbols(repeatingDigit, repeatCount)...)
				doingLeadingZeros = false
			}
		} else {
			symbolsResult = append(symbolsResult, Symbols(repeatingDigit, repeatCount)...)
		}
	}
	return symbolsResult
}

// This symbol is in the "Leading Zeros" alphabet
func LeadingZerosSymbol(count int) types.TSymbol {
	if count > digitsNeeded {
		panic("too many zeros")
	}
	return types.TSymbol(count)
}
func LeadingZerosFrom(s types.TSymbol) int {
	zeros := int(s)
	if zeros > digitsNeeded {
		panic("too many zeros")
	}
	return zeros
}

type MetaDigit struct {
	Digit       int
	RepeatCount int
}

// These symbols are in the "Meta Digit" alphabet, and mean "Repating 9s" or "Repeating 0s"
func RepeatingDigitSymbol(digit int, repeatCount int) types.TSymbol {
	if repeatCount < 1 {
		panic("not enough digits")
	}
	if repeatCount > digitsNeeded {
		panic("too many digits")
	}
	return types.TSymbol(digit + 10*repeatCount)
}
func RepeatingDigitFrom(s types.TSymbol) MetaDigit {
	metaDigit := MetaDigit{int(s % 10), int(s / 10)}
	if metaDigit.RepeatCount < 1 {
		panic("not enough digits")
	}
	if metaDigit.RepeatCount > digitsNeeded {
		panic("too many digits")
	}
	return metaDigit
}
func NeedAnotherMetaDigit(leadingZeros int, metaDigits []MetaDigit) bool {
	if leadingZeros > digitsNeeded {
		panic("too many digits")
	}
	digitsSoFar := leadingZeros
	for _, metaDigit := range metaDigits {
		digitsSoFar += metaDigit.RepeatCount
	}
	if digitsSoFar > digitsNeeded {
		panic("too many digits")
	}
	return digitsSoFar < digitsNeeded
}

// These symbols are also in the "Meta Digit" alphabet
func SingleDigitSymbol(digit int) types.TSymbol {
	if digit < 1 || digit > 8 {
		panic("bad digit")
	}
	return RepeatingDigitSymbol(digit, 1)
}

func Symbols(digit int, repeatCount int) []types.TSymbol {
	if repeatCount < 1 {
		panic("not enough digits")
	}
	if repeatCount > digitsNeeded {
		panic("too many digits")
	}
	if digit == 0 || digit == 9 {
		// A single meta-digit represents between 1 amd 19 zeros
		// A single meta-digit represents between 1 amd 19 nines
		return []types.TSymbol{RepeatingDigitSymbol(digit, repeatCount)}
	} else {
		// Digits 1 through 8 do not have special "repeat symbols"
		// They've come in here as repeats, but need to be repeated explicitly
		result := make([]types.TSymbol, repeatCount)
		for i := range repeatCount {
			result[i] = SingleDigitSymbol(digit)
		}
		return result
	}
}

func EatLeadingRepeatingDigit(inputNumber types.TData, digitCount int, digitsPow10 types.TData) (
	repeatingDigit int, repeatCount int, scraps types.TData, scrapsDigitCount int, remainingDigitsPow10 types.TData) {
	if inputNumber == 0 {
		return 0, digitCount, 0, 0, 0 // Number was all zeroes
	}

	// Scraps counts down from the original number
	scraps = inputNumber
	scrapsDigitCount = digitCount
	remainingDigitsPow10 = digitsPow10

	// First digit
	repeatingDigit = int(scraps / digitsPow10)
	repeatCount = 1 // Until we find otherwise
	if repeatingDigit > 0 {
		scraps = scraps % (types.TData(repeatingDigit) * digitsPow10)
	} else {
		// Zero digit; scraps stays the same
	}
	scrapsDigitCount--
	remainingDigitsPow10 /= 10
	if scraps == 0 {
		return repeatingDigit, repeatCount, scraps, scrapsDigitCount, remainingDigitsPow10
	}

	// Next digit
	nextDigit := int(scraps / remainingDigitsPow10)
	nextScraps := scraps
	if nextDigit > 0 {
		nextScraps = scraps % (types.TData(nextDigit) * remainingDigitsPow10)
	} else {
		// zero digit; nextScraps is just scraps
	}
	nextScrapsDigitCount := scrapsDigitCount - 1
	nextRepeatCount := repeatCount + 1
	nextRemainingDigitsPow10 := remainingDigitsPow10 / 10
	for nextDigit == repeatingDigit {
		scraps = nextScraps
		repeatCount = nextRepeatCount
		scrapsDigitCount = nextScrapsDigitCount
		remainingDigitsPow10 = nextRemainingDigitsPow10

		nextRemainingDigitsPow10 = remainingDigitsPow10 / 10
		if nextRemainingDigitsPow10 == 0 {
			return repeatingDigit, repeatCount, scraps, scrapsDigitCount, remainingDigitsPow10
		}

		nextDigit = int(scraps / remainingDigitsPow10)
		if nextDigit > 0 {
			nextScraps = scraps % (types.TData(nextDigit) * remainingDigitsPow10)
		} else {
			nextScraps = scraps
		}
		nextRepeatCount = repeatCount + 1
	}
	return repeatingDigit, repeatCount, scraps, scrapsDigitCount, remainingDigitsPow10
}

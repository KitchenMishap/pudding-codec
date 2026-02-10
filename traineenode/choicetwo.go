package traineenode

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/types"
)

type ChoiceTwo struct {
	switchNode  scribenode.IScribeNode
	optionNodes [2]scribenode.IBidderScribe

	// Observations
	alphabet        alphabets.Alphabet        // The symbols we've seen
	alphabetProfile alphabets.AlphabetProfile // How often each was seen

	// Metadata (the thing that is learned from observations)
	switchSymbolFromSequence  map[types.TSymbol]types.TSymbol // Indexed by a single-symbol sequence
	sequenceSymbolsFromSwitch [][]types.TSymbol               // Gives a list of single-symbol sequences
}

func NewChoiceTwo(switchNode scribenode.IScribeNode,
	optionNodes [2]scribenode.IBidderScribe) *ChoiceTwo {
	result := ChoiceTwo{}
	result.switchNode = switchNode
	result.optionNodes = optionNodes
	result.alphabet = nil                  // Start with no observations
	result.alphabetProfile = nil           // Start with no observations
	result.switchSymbolFromSequence = nil  // Start untrained
	result.sequenceSymbolsFromSwitch = nil // Start untrained
	return &result
}

func (ct *ChoiceTwo) Observe(sampleSequences [][]types.TSymbol) {
	ct.alphabetProfile, ct.alphabet =
		alphabets.AlphabetProfileFromSampleFirstSymbol(sampleSequences)
}

func (ct *ChoiceTwo) Improve() error {
	alphabetSize := len(ct.alphabet)
	switchSize := 2
	// Forget everything previously learned
	ct.switchSymbolFromSequence = make(map[types.TSymbol]types.TSymbol, alphabetSize)
	ct.sequenceSymbolsFromSwitch = make([][]types.TSymbol, switchSize)
	for i := 0; i < switchSize; i++ {
		ct.sequenceSymbolsFromSwitch[i] = make([]types.TSymbol, 0)
	}
	// For each symbol (first symbol of sequence) seen...
	for _, symbol := range ct.alphabet {
		// Just this once, decide on the appropriate switch symbol to use
		// (the one representing the cheapest of the bids offered
		// by the options for this symbol)
		switchSymbol, refused, err := ct.chooseBestOption([]types.TSymbol{symbol})
		if err != nil {
			return err
		}
		if refused {
			panic("both switch options refused to bid")
		}
		// Learn this switch symbol for this sequence
		ct.switchSymbolFromSequence[symbol] = switchSymbol
		ct.sequenceSymbolsFromSwitch[switchSymbol] =
			append(ct.sequenceSymbolsFromSwitch[switchSymbol], symbol)
	}
	return nil
}

func (ct *ChoiceTwo) Encode(sequence []types.TSymbol, writer bitstream.IBitWriter) (refused bool, err error) {
	if len(sequence) != 1 {
		panic("can only cope with sequences of length one")
	}

	var switchSymbol types.TSymbol
	// If we have learned anything...
	if ct.switchSymbolFromSequence != nil {
		switchSymbol = ct.switchSymbolFromSequence[sequence[0]]
	} else {
		switchSymbol, refused, err = ct.chooseBestOption(sequence)
		if err != nil {
			return false, err
		}
		if refused {
			return true, nil
		}
	}

	// Encode switch
	switchSequence := []types.TSymbol{switchSymbol} // Sequence of 1 symbol
	refused, err = ct.switchNode.Encode(switchSequence, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ChoiceTwo: switch refused to encode")
	}

	// Encode sequence
	refused, err = ct.optionNodes[switchSymbol].Encode(sequence, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ChoiceTwo: option node refused to encode")
	}

	return false, nil
}

func (ct *ChoiceTwo) Decode(reader bitstream.IBitReader) ([]types.TSymbol, error) {
	// Read the switch
	switchSequence, err := ct.switchNode.Decode(reader)
	if err != nil {
		return []types.TSymbol{}, err
	}
	if len(switchSequence) != 1 {
		panic("wasn't expecting multiple switch settings")
	}
	switchSymbol := switchSequence[0]

	return ct.optionNodes[switchSymbol].Decode(reader)
}

func (ct *ChoiceTwo) chooseBestOption(sequence []types.TSymbol) (
	switchSymbol types.TSymbol, refused bool, err error) {
	// Assess
	cost0, refuse0, err := ct.optionNodes[0].BidBits(sequence)
	if err != nil {
		return types.TSymbol(0), false, err
	}
	cost1, refuse1, err := ct.optionNodes[1].BidBits(sequence)
	if err != nil {
		return types.TSymbol(0), false, err
	}

	// Choose
	// If both choices refused, WE refuse
	if refuse0 && refuse1 {
		return types.TSymbol(0), true, nil
	}
	switchSymbol = types.TSymbol(0)
	if refuse0 || cost1 < cost0 {
		switchSymbol = types.TSymbol(1)
	}

	return switchSymbol, false, nil
}

const switchPositions = 2
const switchBits = 1
const dataBits = 64

func (ct *ChoiceTwo) EncodeMyMetaData(writer bitstream.IBitWriter) error {
	switchScribe := scribenode.NewFixedBits(switchBits)
	dataScribe := scribenode.NewFixedBits(dataBits)
	refused, err := dataScribe.Encode([]types.TSymbol{types.TSymbol(len(ct.switchSymbolFromSequence))}, writer)
	if err != nil {
		return err
	}
	if refused {
		panic("scribe refused to encode ChoiceTwo count metadata")
	}
	for dataSymbol, switchSymbol := range ct.switchSymbolFromSequence {
		refused, err = dataScribe.Encode([]types.TSymbol{dataSymbol}, writer)
		if err != nil {
			return err
		}
		if refused {
			panic("scribe refused to encode ChoiceTwo data metadata")
		}
		refused, err = switchScribe.Encode([]types.TSymbol{switchSymbol}, writer)
		if err != nil {
			return err
		}
		if refused {
			panic("scribe refused to encode ChoiceTwo switch metadata")
		}
	}
	// Tell the children to encode their metadata
	// ... No need as they're all scribes (no metadata)
	return nil
}

func (ct *ChoiceTwo) DecodeMyMetaData(reader bitstream.IBitReader) error {
	switchScribe := scribenode.NewFixedBits(switchBits)
	dataScribe := scribenode.NewFixedBits(dataBits)
	counts, err := dataScribe.Decode(reader)
	if err != nil {
		return err
	}
	if len(counts) != 1 {
		panic("didn't expect multiple counts")
	}
	count := counts[0]

	ct.sequenceSymbolsFromSwitch = make([][]types.TSymbol, switchPositions)
	for i := 0; i < switchPositions; i++ {
		ct.sequenceSymbolsFromSwitch[i] = make([]types.TSymbol, 0)
	}
	ct.switchSymbolFromSequence = make(map[types.TSymbol]types.TSymbol, count)

	for range count {
		dataSymbolSequence, err := dataScribe.Decode(reader)
		if err != nil {
			return err
		}
		if len(dataSymbolSequence) != 1 {
			panic("expecting sequence of data symbols to be length 1")
		}
		dataSymbol := dataSymbolSequence[0]
		switchSymbolSequence, err := switchScribe.Decode(reader)
		if err != nil {
			return err
		}
		if len(switchSymbolSequence) != 1 {
			panic("expecting sequence of switch symbols to be length 1")
		}
		switchSymbol := switchSymbolSequence[0]
		ct.sequenceSymbolsFromSwitch[int(switchSymbol)] =
			append(ct.sequenceSymbolsFromSwitch[int(switchSymbol)], dataSymbol)
		ct.switchSymbolFromSequence[dataSymbol] = switchSymbol
	}
	// Tell the children to decode their metadata
	// ... No need as they're all scribes (no metadata)
	return nil
}

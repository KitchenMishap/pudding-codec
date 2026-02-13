package traineenode

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-codec/types"
)

type ShannonFano struct {
	switchNode  scribenode.IScribeNode
	optionNodes [2]scribenode.IBidderScribe

	factory ShannonFanoFactory // The builder for recursive children

	// Observations
	alphabetCounts *alphabets.AlphabetCount // The symbols we've observed

	// Metadata (the thing that is learned from observations)
	switchSymbolFromSequence  map[types.TSymbol]types.TSymbol // Indexed by a single-symbol sequence
	sequenceSymbolsFromSwitch [][]types.TSymbol               // Gives a list of single-symbol sequences
}

// Check that implements
var _ ITraineeNode = (*ShannonFano)(nil)

type ShannonFanoFactory func() *ShannonFano

func NewRecursiveShannonFano(switchNode scribenode.IScribeNode) *ShannonFano {
	var sf *ShannonFano

	// The factory is a closure that knows how to make "more of me"
	factory := func() *ShannonFano {
		// Each sub-node needs its own switchNode instance!
		// (Assuming you want a 1-bit fixed switch for the whole tree)
		newSwitch := scribenode.NewFixedBits(1)
		return NewRecursiveShannonFano(newSwitch)
	}

	sf = &ShannonFano{
		switchNode:     switchNode,
		factory:        factory,
		alphabetCounts: alphabets.NewAlphabetCount(10),
	}
	return sf
}

func (sf *ShannonFano) Observe(samples []types.TSymbol) error {
	sf.alphabetCounts.AddData(samples)
	return nil
}

func (sf *ShannonFano) Improve() error {
	if sf.alphabetCounts.Size() == 0 {
		return nil
	}
	alphabetProfile, _ := sf.alphabetCounts.MakeAlphabetProfile()
	// All we need to do is kick off the recursion
	err := sf.ImproveRecursive(alphabetProfile)
	sf.alphabetCounts.Reset() // Clear observations
	return err
}

func (ct *ShannonFano) ImproveRecursive(profile alphabets.AlphabetProfile) error {
	leftProfile, rightProfile := alphabets.SplitProfile(profile)

	// Initialize map for this node
	ct.switchSymbolFromSequence = make(map[types.TSymbol]types.TSymbol, len(profile))

	// 1. Populate the map for the LEFT side
	for _, sc := range leftProfile {
		ct.switchSymbolFromSequence[sc.Symbol] = 0
	}
	// Set up the LEFT node
	if len(leftProfile) == 1 {
		ct.optionNodes[0] = scribenode.NewLiteralScribe(leftProfile[0].Symbol)
	} else if len(leftProfile) > 1 {
		child := ct.factory()
		if err := child.ImproveRecursive(leftProfile); err != nil {
			return err
		}
		ct.optionNodes[0] = child
	} else {
		// Empty side. This can occur if there's only one symbol in the alphabet
		ct.optionNodes[0] = scribenode.NewLiteralScribe(0) // Dummy
	}

	// 2. Populate the map for the RIGHT side
	for _, sc := range rightProfile {
		ct.switchSymbolFromSequence[sc.Symbol] = 1
	}
	// Set up the RIGHT node
	if len(rightProfile) == 1 {
		ct.optionNodes[1] = scribenode.NewLiteralScribe(rightProfile[0].Symbol)
	} else if len(rightProfile) > 1 {
		child := ct.factory()
		if err := child.ImproveRecursive(rightProfile); err != nil {
			return err
		}
		ct.optionNodes[1] = child
	} else {
		// Empty side. This can occur if there's only one symbol in the alphabet
		ct.optionNodes[1] = scribenode.NewLiteralScribe(0) // Dummy
	}

	return nil
}

func (ct *ShannonFano) DecodeMyMetaData(reader bitstream.IBitReader) error {
	typeScribe := scribenode.NewFixedBits(1)
	dataScribe := scribenode.NewFixedBits(64)

	for i := 0; i < 2; i++ {
		// 1. Read the "Control Bit" (Branch vs Leaf)
		typeSymbol, err := typeScribe.Decode(reader)
		if err != nil {
			return err
		}
		isBranch := typeSymbol == 1

		if isBranch {
			// 2. It's a branch! Spawn a child and recurse
			child := ct.factory()
			err := child.DecodeMyMetaData(reader)
			if err != nil {
				return err
			}
			ct.optionNodes[i] = child
		} else {
			// 3. It's a leaf! Read the symbol and create a LiteralScribe
			symbol, err := dataScribe.Decode(reader)
			if err != nil {
				return err
			}
			ct.optionNodes[i] = scribenode.NewLiteralScribe(symbol)
		}
	}
	return nil
}

func (ct *ShannonFano) chooseBestOption(symbol types.TSymbol) (
	switchSymbol types.TSymbol, refused bool, err error) {
	// Assess
	cost0, refuse0, err := ct.optionNodes[0].BidBits(symbol)
	if err != nil {
		return 0, false, err
	}
	cost1, refuse1, err := ct.optionNodes[1].BidBits(symbol)
	if err != nil {
		return 0, false, err
	}

	// Choose
	// If both choices refused, WE refuse
	if refuse0 && refuse1 {
		return 0, true, nil
	}
	switchSymbol = types.TSymbol(0)
	if refuse0 || cost1 < cost0 {
		switchSymbol = types.TSymbol(1)
	}

	return switchSymbol, false, nil
}

func (ct *ShannonFano) EncodeMyMetaData(writer bitstream.IBitWriter) error {
	typeScribe := scribenode.NewFixedBits(1)
	dataScribe := scribenode.NewFixedBits(64)

	for i := 0; i < 2; i++ {
		// Use a type switch or assertion to see what the child is
		if child, ok := ct.optionNodes[i].(*ShannonFano); ok {
			// 1. Write the "Branch" flag
			refused, err := typeScribe.Encode(1, writer)
			if err != nil {
				return err
			}
			if refused {
				panic("refused to write the branch flag")
			}
			// 2. Recurse!
			err = child.EncodeMyMetaData(writer)
			if err != nil {
				return err
			}
		} else if leaf, ok := ct.optionNodes[i].(*scribenode.LiteralScribe); ok {
			// 1. Write the "Leaf" flag
			refused, err := typeScribe.Encode(0, writer)
			if err != nil {
				return err
			}
			if refused {
				panic("refused to write the leaf flag")
			}
			// 2. Write the symbol (You'll need GetSymbol() on LiteralScribe)
			refused, err = dataScribe.Encode(leaf.GetSymbol(), writer)
			if err != nil {
				return err
			}
			if refused {
				panic("literal scribe refused to write the symbol")
			}
		}
	}
	return nil
}

func (ct *ShannonFano) Encode(symbol types.TSymbol, writer bitstream.IBitWriter) (refused bool, err error) {
	var switchSymbol types.TSymbol
	// If we have learned anything...
	if ct.switchSymbolFromSequence != nil {
		switchSymbol = ct.switchSymbolFromSequence[symbol]
	} else {
		switchSymbol, refused, err = ct.chooseBestOption(symbol)
		if err != nil {
			return false, err
		}
		if refused {
			return true, nil
		}
	}

	// Encode switch
	refused, err = ct.switchNode.Encode(switchSymbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ShannonFano: switch refused to encode")
	}

	// Encode symbol
	refused, err = ct.optionNodes[switchSymbol].Encode(symbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		panic("ShannonFano: option node refused to encode")
	}

	return false, nil
}

func (ct *ShannonFano) Decode(reader bitstream.IBitReader) (types.TSymbol, error) {
	// Read the switch
	switchSymbol, err := ct.switchNode.Decode(reader)
	if err != nil {
		return 0, err
	}
	return ct.optionNodes[switchSymbol].Decode(reader)
}

func (ct *ShannonFano) BidBits(symbol types.TSymbol) (types.TBitCount, bool, error) {
	// 1. Determine which branch this symbol belongs to
	// (We use our learned mapping for speed)
	switchSymbol, exists := ct.switchSymbolFromSequence[symbol]
	if !exists {
		// If we haven't learned this symbol, we can't bid on it
		return 0, true, nil
	}

	// 2. Ask that specific child for its bit cost
	childCost, refused, err := ct.optionNodes[switchSymbol].BidBits(symbol)
	if err != nil || refused {
		return 0, refused, err
	}

	// 3. Total cost = 1 bit (for our switch) + whatever the child needs
	// (In your architecture, the switchNode is FixedBits(1))
	return 1 + childCost, false, nil
}

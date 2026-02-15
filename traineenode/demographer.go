package traineenode

import (
	"github.com/KitchenMishap/pudding-codec/alphabets"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
	"math"
)

// Demographer is a trainee node which trains two or more children that have different skill sets.
// It then measures their aptitude, and trains them a SECOND time, encouraging specialization.
// More detail:
// In its Observe() phase, it trains the children a first time on equal data.
// In its Improve() phase, it assesses them on the original data, but for each datum, it selects
// the best performing child for that datum. It sends the datum to the best performing child via Observe().
// Still in its Improve() phase, it then asks all the children to Improve() on this second, filtered
// observation set.
type Demographer struct {
	switchRepresentationNode ITraineeNode
	optionNodes              []ITraineeNode

	// Observations
	alphabetCounts *alphabets.AlphabetCount // The symbols we've observed

	// No metadata (except in children nodes)
}

func NewDemographer(switchRepresentationNode ITraineeNode,
	optionNodes []ITraineeNode) ITraineeNode {
	result := Demographer{}
	result.switchRepresentationNode = switchRepresentationNode
	result.optionNodes = optionNodes
	result.alphabetCounts = alphabets.NewAlphabetCount(1000)
	return &result
}

// Check that implements
var _ ITraineeNode = (*Demographer)(nil)

func (dm *Demographer) Observe(samples []types.TSymbol) error {
	// Store here
	dm.alphabetCounts.AddData(samples)
	// Send to children
	for _, node := range dm.optionNodes {
		err := node.Observe(samples)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dm *Demographer) Improve() error {
	// 1) Tell the children to improve based on initial (unfiltered) observations
	//		(This also tells the children to reset those observation stores)
	// (Not the switch node - it hasn't been sent any Observe() calls yet)
	for _, node := range dm.optionNodes {
		err := node.Improve()
		if err != nil {
			return err
		}
	}

	// 2) For each datum, select the best performing child and tell it to observe that data
	alphabetProfile, _ := dm.alphabetCounts.MakeAlphabetProfile()
	for _, entry := range alphabetProfile {
		datum := entry.Symbol
		count := entry.Count

		cheapestBid := uint64(math.MaxUint64)
		cheapestBidder := -1
		for i, node := range dm.optionNodes {
			bid, refused, err := node.BidBits(datum)
			if err != nil {
				return err
			}
			if !refused {
				if bid < cheapestBid {
					cheapestBid = bid
					cheapestBidder = i
				}
			}
		} // for child node

		// Re-train the "best" child for this data on this (filtered) data
		weightedSamples := make([]types.TSymbol, count)
		for i := range weightedSamples {
			weightedSamples[i] = datum
		}

		err := dm.optionNodes[cheapestBidder].Observe(weightedSamples)
		if err != nil {
			return err
		}

		// Train the switch too
		switchSymbol := types.TSymbol(cheapestBidder)
		weightedSwitch := make([]types.TSymbol, count)
		for i := range weightedSwitch {
			weightedSwitch[i] = switchSymbol
		}
		err = dm.switchRepresentationNode.Observe(weightedSwitch)
		if err != nil {
			return err
		}
	} // for datum

	// 3) Tell all the children to retrain on the new set of data they've observed
	for _, node := range dm.optionNodes {
		err := node.Improve()
		if err != nil {
			return err
		}
	}
	// And the switch too
	err := dm.switchRepresentationNode.Improve()
	if err != nil {
		return err
	}

	// AND RESET MY ALPHABET COUNT!
	dm.alphabetCounts.Reset()

	return nil
}

func (dm *Demographer) EncodeMyMetaData(writer bitstream.IBitWriter) error {
	// No metadata of our own; just our children's
	err := dm.switchRepresentationNode.EncodeMyMetaData(writer)
	if err != nil {
		return err
	}
	for _, node := range dm.optionNodes {
		err = node.EncodeMyMetaData(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dm *Demographer) DecodeMyMetaData(reader bitstream.IBitReader) error {
	// No metadata of our own; just our children's
	err := dm.switchRepresentationNode.DecodeMyMetaData(reader)
	if err != nil {
		return err
	}
	for _, node := range dm.optionNodes {
		err = node.DecodeMyMetaData(reader)
		if err != nil {
			return err
		}
	}
	return nil
}
func (dm *Demographer) Encode(symbol types.TSymbol,
	writer bitstream.IBitWriter) (bool, error) {

	// We have to try all the options to see which is cheapest
	cheapestBid := uint64(math.MaxUint64)
	cheapestBidder := -1
	for i, node := range dm.optionNodes {
		bid, refused, err := node.BidBits(symbol)
		if err != nil {
			return false, err
		}
		if !refused {
			if bid < cheapestBid {
				cheapestBid = bid
				cheapestBidder = i
			}
		}
	} // for child node

	// Store the switch setting (needed by Decode())
	switchSymbol := types.TSymbol(cheapestBidder)
	refused, err := dm.switchRepresentationNode.Encode(switchSymbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		return false, nil
	}

	// Store the specific encoding
	refused, err = dm.optionNodes[cheapestBidder].Encode(symbol, writer)
	if err != nil {
		return false, err
	}
	if refused {
		return false, nil
	}

	return false, nil
}
func (dm *Demographer) Decode(reader bitstream.IBitReader) (types.TSymbol, error) {
	// Decode the switch setting
	switchSymbol, err := dm.switchRepresentationNode.Decode(reader)
	if err != nil {
		return 0, err
	}

	// Decode the data using the chosen option
	return dm.optionNodes[switchSymbol].Decode(reader)
}
func (dm *Demographer) BidBits(symbol types.TSymbol) (types.TBitCount, bool, error) {
	counter := bitstream.NewBitCounter()
	refused, err := dm.Encode(symbol, counter)
	if err != nil {
		return 0, false, err
	}
	if refused {
		return 0, true, nil
	}
	return counter.CountBits(), false, nil
}

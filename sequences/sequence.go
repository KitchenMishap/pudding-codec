package sequences

import "github.com/KitchenMishap/pudding-codec/types"

// A sequence is generally thought of as a message.
// A sequence is a slice of symbols; each slice generally is a member of a DIFFERENT alphabet.
// (The symbol TSymbol{23} generally means a different symbol to TSymbol{23} appearing
// in a different position in a sequence!)

// A sequence sample is a slice of sequences.
// Sequences are very often OF LENGTH ONE. Much code here ASSUMES THIS.
// Here are some functions that deal with sequences that are/aren't of length one.

func SequenceAsSingleSymbol(sequence []types.TSymbol) types.TSymbol {
	if len(sequence) != 1 {
		panic("expecting sequence to be length one")
	}
	return sequence[0]
}

func SliceOfSingleSymbolsFromSampleOfSequences(sampleSequences [][]types.TSymbol) []types.TSymbol {
	result := make([]types.TSymbol, len(sampleSequences))
	for i, seq := range sampleSequences {
		result[i] = SequenceAsSingleSymbol(seq)
	}
	return result
}

func SingleSymbolDataToSampleOfSequences(data []types.TData) [][]types.TSymbol {
	result := make([][]types.TSymbol, len(data))
	for i, datapoint := range data {
		result[i] = []types.TSymbol{datapoint}
	}
	return result
}

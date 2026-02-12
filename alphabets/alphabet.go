package alphabets

import (
	"github.com/KitchenMishap/pudding-codec/types"
	"sort"
)

type DataSet = []types.TData

type SymbolCount struct {
	Symbol types.TSymbol
	Count  types.TCount
}

type AlphabetProfile = []SymbolCount

type Alphabet = []types.TSymbol

// Each element can be from a DIFFERENT Alphabet (implied by context)
type Sentence = []types.TSymbol

type AlphabetCount struct {
	capGuess int
	counts   map[types.TSymbol]types.TCount
}

func NewAlphabetCount(sizeGuess int) *AlphabetCount {
	result := AlphabetCount{}
	result.capGuess = sizeGuess
	result.counts = make(map[types.TSymbol]types.TCount, sizeGuess)
	return &result
}

func (ac *AlphabetCount) AddData(dataSet DataSet) {
	for _, v := range dataSet {
		ac.counts[v]++
	}
}

func (ac *AlphabetCount) Size() int { return len(ac.counts) }

func (ac *AlphabetCount) Reset() {
	ac.counts = make(map[types.TSymbol]types.TCount, ac.capGuess)
}

func (ac *AlphabetCount) MakeAlphabetProfile() (AlphabetProfile, Alphabet) {
	// Turn map into an unsorted slice
	favourites := make(AlphabetProfile, len(ac.counts))
	i := 0
	for k, v := range ac.counts {
		favourites[i].Symbol = k
		favourites[i].Count = v
		i++
	}
	// Sort slice (biggest Count first)
	sort.Slice(favourites, func(i, j int) bool {
		return favourites[i].Count > favourites[j].Count
	})
	// Create list of unique values
	uniques := make(Alphabet, len(favourites))
	for i, v := range favourites {
		uniques[i] = v.Symbol
	}
	return favourites, uniques
}

func AlphabetProfileFromData(dataSet DataSet) (AlphabetProfile, Alphabet) {
	counts := make(map[types.TData]types.TCount, 1000)
	for _, v := range dataSet {
		counts[v]++
	}
	// Turn map into an unsorted slice
	favourites := make(AlphabetProfile, len(counts))
	i := 0
	for k, v := range counts {
		favourites[i].Symbol = k
		favourites[i].Count = v
		i++
	}
	// Sort slice (biggest Count first)
	sort.Slice(favourites, func(i, j int) bool {
		return favourites[i].Count > favourites[j].Count
	})
	// Create list of unique values
	uniques := make(Alphabet, len(favourites))
	for i, v := range favourites {
		uniques[i] = v.Symbol
	}
	return favourites, uniques
}

func AlphabetProfileFromSampleFirstSymbol(samples [][]types.TSymbol) (AlphabetProfile, Alphabet) {
	counts := make(map[types.TData]types.TCount, 100)
	for _, sequence := range samples {
		if len(sequence) != 1 {
			panic("can only cope with single symbol sequences")
		}
		counts[sequence[0]]++
	}
	// Turn map into an unsorted slice
	favourites := make(AlphabetProfile, len(counts))
	i := 0
	for k, v := range counts {
		favourites[i].Symbol = k
		favourites[i].Count = v
		i++
	}
	// Sort slice (biggest Count first)
	sort.Slice(favourites, func(i, j int) bool {
		return favourites[i].Count > favourites[j].Count
	})
	// Create list of unique values
	uniques := make(Alphabet, len(favourites))
	for i, v := range favourites {
		uniques[i] = v.Symbol
	}
	return favourites, uniques
}

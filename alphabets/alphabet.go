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

func NewAlphabet(symbolCount types.TCount) Alphabet {
	result := make(Alphabet, symbolCount)
	for i := range symbolCount {
		result[i] = types.TSymbol(i)
	}
	return result
}

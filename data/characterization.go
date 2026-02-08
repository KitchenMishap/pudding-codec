package data

import (
	"github.com/KitchenMishap/pudding-codec/types"
	"sort"
)

type DataSet = []types.TData

type ValueCount struct {
	Value types.TData
	Count types.TData
}

type ValueFavourites = []ValueCount

type UniqueValues = []types.TData

func FavouritesUniques(dataSet DataSet) (ValueFavourites, UniqueValues) {
	counts := make(map[types.TData]types.TCount, 1000)
	for _, v := range dataSet {
		counts[v]++
	}
	// Turn map into an unsorted slice
	favourites := make(ValueFavourites, len(counts))
	i := 0
	for k, v := range counts {
		favourites[i].Value = k
		favourites[i].Count = v
		i++
	}
	// Sort slice (biggest Count first)
	sort.Slice(favourites, func(i, j int) bool {
		return favourites[i].Count > favourites[j].Count
	})
	// Create list of unique values
	uniques := make(UniqueValues, len(favourites))
	for i, v := range favourites {
		uniques[i] = v.Value
	}
	return favourites, uniques
}

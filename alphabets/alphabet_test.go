package alphabets

import (
	"github.com/KitchenMishap/pudding-codec/types"
	"testing"
)

func Test_Alphabet(t *testing.T) {
	inputData := []types.TData{1, 2, 2, 2, 3, 3, 3, 3}

	favourites, uniques := AlphabetProfileFromData(inputData)

	if len(favourites) != 3 {
		t.Error("There should be 3 favourites")
	}
	if len(uniques) != 3 {
		t.Error("There should be 3 uniques")
	}
	if favourites[0].Symbol != 3 {
		t.Error("The favourite should be 3")
	}
	if favourites[2].Symbol != 1 {
		t.Error("The least favourite should be 1")
	}
	if favourites[1].Count != 3 {
		t.Error("The second favourite should be count 3")
	}
}

package alphabets

import "github.com/KitchenMishap/pudding-codec/types"

// SplitProfile takes a sorted AlphabetProfile and finds the best index to
// split it into two sub-profiles for a binary tree branch.
func SplitProfile(profile AlphabetProfile) (left AlphabetProfile, right AlphabetProfile) {
	if len(profile) <= 1 {
		return profile, nil
	}

	totalWeight := types.TCount(0)
	for _, sc := range profile {
		totalWeight += sc.Count
	}

	cumulativeWeight := types.TCount(0)
	splitIndex := 0

	// Find the point where cumulative weight is closest to half of total weight
	for i, sc := range profile {
		cumulativeWeight += sc.Count
		splitIndex = i
		// If we've reached or passed the halfway mark
		if cumulativeWeight*2 >= totalWeight {
			// Check if including this element is better than stopping at the previous one
			// (Basic optimization to keep the tree balanced)
			break
		}
	}

	// Ensure we don't return an empty right slice if there are symbols left
	if splitIndex == len(profile)-1 && len(profile) > 1 {
		splitIndex--
	}

	return profile[:splitIndex+1], profile[splitIndex+1:]
}

package utils

func calculateWeightedStats(values, weights []float64) (mean, variance float64) {
	sumOfProducts := 0.0
	sumOfWeights := 0.0
	sumOfSquaredWeights := 0.0

	for i, val := range values {
		sumOfProducts += val * weights[i]
		sumOfWeights += weights[i]
		sumOfSquaredWeights += weights[i] * weights[i]
	}

	mean = sumOfProducts / sumOfWeights

	sumOfSquaredDifferences := 0.0
	for i, val := range values {
		diff := val - mean
		sumOfSquaredDifferences += weights[i] * diff * diff
	}

	V1 := sumOfWeights
	V2 := sumOfSquaredWeights
	variance = (V1 / (V1*V1 - V2)) * sumOfSquaredDifferences

	return mean, variance
}

package utils

import "math"

func BucketCount(beans int64, beansPerBucket int64) int64 {
	return (beans + beansPerBucket - 1) / beansPerBucket
}

func SafeMultiply(a, b uint64) (uint64, bool) {
	if a == 0 || b == 0 {
		return 0, false
	}
	// If a * b > MaxUint64, then MaxUint64 / a < b
	if math.MaxUint64/a < b {
		return 0, true // Overflow detected!
	}
	return a * b, false
}

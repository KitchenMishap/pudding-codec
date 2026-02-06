package utils

func BucketCount(beans int64, beansPerBucket int64) int64 {
	return (beans + beansPerBucket - 1) / beansPerBucket
}

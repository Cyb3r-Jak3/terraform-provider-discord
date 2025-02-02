package utils

import "hash/crc32"

func Hashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func Contains[T comparable](array []T, value T) bool {
	for _, elem := range array {
		if elem == value {
			return true
		}
	}

	return false
}

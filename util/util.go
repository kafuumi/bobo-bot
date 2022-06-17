package util

import (
	"fmt"
)

func IsError(err error, info string) bool {
	if err != nil {
		fmt.Printf("[error] %s, %v\n", info, err)
		return true
	}
	return false
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func SliceSet[T any](s []T, i int, item T) []T {
	s = sliceCheck(s, i)
	s[i] = item
	return s
}

func SliceGet[T any](s []T, i int) ([]T, T) {
	s = sliceCheck(s, i)
	return s, s[i]
}

func sliceCheck[T any](s []T, i int) []T {
	l := len(s) - 1
	var zero T
	if i > l {
		for ; l < i; l++ {
			s = append(s, zero)
		}
	}
	return s
}

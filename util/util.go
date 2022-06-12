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

// Condition 三目运算符
func Condition[T any](con bool, a, b T) T {
	if con {
		return a
	}
	return b
}

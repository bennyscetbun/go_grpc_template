package random

import (
	"math/rand"
)

const numberBytes = "0123456789"

func RandNumber(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = numberBytes[rand.Intn(len(numberBytes))]
	}
	return string(b)
}

const stringBytes = "0123456789abcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = stringBytes[rand.Intn(len(stringBytes))]
	}
	return string(b)
}

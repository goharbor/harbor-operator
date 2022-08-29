package test

import (
	"fmt"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))] //nolint:gosec
	}

	return string(b)
}

const prefixLength = 8

func NewName(prefix string) string {
	return fmt.Sprintf("test-%s-%s", prefix, randStringRunes(prefixLength))
}

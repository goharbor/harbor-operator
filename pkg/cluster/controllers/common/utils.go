package common

import (
	"bytes"
	"math/rand"
	"strings"
	"time"
)

func Bools(b bool) *bool {
	return &b
}

// RandomString returns random string.
func RandomString(randLength int, randType string) (result string) {
	var (
		num   = "0123456789"
		lower = "abcdefghijklmnopqrstuvwxyz"
		upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	)

	b := bytes.Buffer{}

	switch {
	case strings.Contains(randType, "0"):
		b.WriteString(num)
	case strings.Contains(randType, "A"):
		b.WriteString(upper)
	default:
		b.WriteString(lower)
	}

	str := b.String()
	strLen := len(str)

	rand.Seed(time.Now().UnixNano())

	b = bytes.Buffer{}

	for i := 0; i < randLength; i++ {
		b.WriteByte(str[rand.Intn(strLen)])
	}

	return b.String()
}

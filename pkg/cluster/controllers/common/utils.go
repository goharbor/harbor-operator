package common

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"strings"
)

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

	b = bytes.Buffer{}

	for i := 0; i < randLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(strLen)))
		if err != nil {
			panic(err)
		}

		b.WriteByte(str[int32(n.Int64())])
	}

	return b.String()
}

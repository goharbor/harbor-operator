package strings

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	charset                = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	nameLen                = 6
	NormalizationSeparator = "-"
	baseInt10              = 10
	baseBitSize            = 64
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

func NormalizeName(name string, suffixes ...string) string {
	if len(suffixes) > 0 {
		name += fmt.Sprintf("%s%s", NormalizationSeparator, strings.Join(suffixes, NormalizationSeparator))
	}

	return name
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

// RandomName generates random names.
func RandomName(prefix string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", prefix, stringWithCharset(nameLen, charset)))
}

// ExtractID extracts ID from location of response.
func ExtractID(location string) (int64, error) {
	idstr := location[strings.LastIndex(location, "/")+1:]

	return strconv.ParseInt(idstr, baseInt10, baseBitSize)
}

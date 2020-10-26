package common

import (
	"bytes"
	"math/rand"
	"strings"
	"time"
)

const (
	UpperStringRandomType = "A"
	LowerStringRandomType = "a"
	NumberRandomType      = "0"
)

func RandomString(randLength int, randType string) (result string) {
	var num = "0123456789"
	var lower = "abcdefghijklmnopqrstuvwxyz"
	var upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := bytes.Buffer{}
	if strings.Contains(randType, "0") {
		b.WriteString(num)
	} else if strings.Contains(randType, "A") {
		b.WriteString(upper)
	} else {
		b.WriteString(lower)
	}

	var str = b.String()
	var strLen = len(str)

	rand.Seed(time.Now().UnixNano())
	b = bytes.Buffer{}
	for i := 0; i < randLength; i++ {
		b.WriteByte(str[rand.Intn(strLen)])
	}
	result = b.String()
	return
}

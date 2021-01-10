package common

import (
	"bytes"
	"math/rand"
	"strings"
	"time"

	"github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
)

func GetIngressPath(ingress v1alpha1.IngressController) (string, error) {
	switch ingress {
	case v1alpha1.IngressControllerDefault:
		return "/", nil
	case v1alpha1.IngressControllerGCE:
		return "/*", nil
	case v1alpha1.IngressControllerNCP:
		return "/.*", nil
	default:
		return "", harbor.ErrInvalidIngressController{Controller: ingress}
	}
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

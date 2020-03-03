package core

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	secretKey = "secretKey"

	keyLength   = 16
	numDigits   = 5
	numSymbols  = 0
	noUpper     = false
	allowRepeat = true
)

func (r *Reconciler) GetSecret(ctx context.Context, core *goharborv1alpha2.Core) (*corev1.Secret, error) {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-core", core.GetName()),
			Namespace: core.GetNamespace(),
		},
		StringData: map[string]string{
			"secret":  password.MustGenerate(keyLength, numDigits, numSymbols, noUpper, allowRepeat),
			secretKey: password.MustGenerate(keyLength, numDigits, numSymbols, noUpper, allowRepeat),
		},
	}, nil
}

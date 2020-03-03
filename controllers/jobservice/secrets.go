package jobservice

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	secretKey = "secret"

	keyLength   = 32
	numDigits   = 10
	numSymbols  = 10
	noUpper     = false
	allowRepeat = true
)

func (r *Reconciler) GetSecrets(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*corev1.Secret, error) {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jobservice", jobservice.GetName()),
			Namespace: jobservice.GetNamespace(),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			secretKey: password.MustGenerate(keyLength, numDigits, numSymbols, noUpper, allowRepeat),
		},
	}, nil
}

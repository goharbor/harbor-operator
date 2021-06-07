package jobservice

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretKey = "secret"

	keyLength   = 32
	numDigits   = 10
	numSymbols  = 10
	noUpper     = false
	allowRepeat = true
)

func (r *Reconciler) GetSecrets(ctx context.Context, jobservice *goharborv1.JobService) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, jobservice.GetName())
	namespace := jobservice.GetNamespace()

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			secretKey: password.MustGenerate(keyLength, numDigits, numSymbols, noUpper, allowRepeat),
		},
	}, nil
}

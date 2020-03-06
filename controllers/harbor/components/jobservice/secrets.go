package jobservice

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	keyLength = 32
	secretKey = "secret"
)

func (j *JobService) GetSecrets(ctx context.Context) []*corev1.Secret {
	operatorName := application.GetName(ctx)
	harborName := j.harbor.Name

	return []*corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      j.harbor.NormalizeComponentName(containerregistryv1alpha1.JobServiceName),
				Namespace: j.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.JobServiceName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				secretKey: password.MustGenerate(keyLength, 10, 10, false, true),
			},
		},
	}
}

func (j *JobService) GetSecretsCheckSum() string {
	// TODO get generation of the secrets
	value := j.harbor.Spec.Components.JobService.RedisSecret
	sum := sha256.New().Sum([]byte(value))

	return fmt.Sprintf("%x", sum)
}

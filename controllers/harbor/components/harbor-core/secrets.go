package core

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
	keyLength = 16
	secretKey = "secretKey"
)

func (c *HarborCore) GetSecrets(ctx context.Context) []*corev1.Secret {
	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.CoreName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			StringData: map[string]string{
				"secret":  password.MustGenerate(keyLength, 5, 0, false, true),
				secretKey: password.MustGenerate(keyLength, 5, 0, false, true),
			},
		},
	}
}

func (c *HarborCore) GetSecretCheckSum() string {
	h := sha256.New()
	return fmt.Sprintf("%x", h.Sum([]byte(c.harbor.Spec.PublicURL)))
}

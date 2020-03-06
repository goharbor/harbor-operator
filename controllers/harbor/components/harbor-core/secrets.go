package core

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
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
				Name:      c.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.CoreName,
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

func (c *HarborCore) GetSecretsCheckSum() string {
	// TODO get generation of the secrets
	value := fmt.Sprintf("%s\n%s", c.harbor.Spec.Components.Core.DatabaseSecret, c.harbor.Spec.Components.Registry.CacheSecret)
	sum := sha256.New().Sum([]byte(value))

	return fmt.Sprintf("%x", sum)
}

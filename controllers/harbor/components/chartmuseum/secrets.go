package chartmuseum

import (
	"context"
	"crypto/sha256"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func (*ChartMuseum) GetSecrets(ctx context.Context) []*corev1.Secret {
	return []*corev1.Secret{}
}

func (c *ChartMuseum) GetSecretsCheckSum() string {
	// TODO get generation of the secrets
	value := fmt.Sprintf("%s\n%s", c.harbor.Spec.Components.ChartMuseum.CacheSecret, c.harbor.Spec.Components.ChartMuseum.StorageSecret)
	sum := sha256.New().Sum([]byte(value))

	return fmt.Sprintf("%x", sum)
}

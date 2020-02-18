package clair

import (
	"context"
	"crypto/sha256"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func (c *Clair) GetSecrets(ctx context.Context) []*corev1.Secret {
	return []*corev1.Secret{}
}

func (c *Clair) GetSecretsCheckSum() string {
	// TODO get generation of the secret
	value := c.harbor.Spec.Components.Clair.DatabaseSecret
	sum := sha256.New().Sum([]byte(value))

	return fmt.Sprintf("%x", sum)
}

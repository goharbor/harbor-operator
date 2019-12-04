package clair

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

func (c *Clair) GetSecrets(ctx context.Context) []*corev1.Secret {
	return []*corev1.Secret{}
}

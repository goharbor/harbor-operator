package registry

import (
	"context"

	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	keyLength = 15
)

func (r *Registry) GetSecrets(ctx context.Context) []*corev1.Secret {
	operatorName := application.GetName(ctx)
	harborName := r.harbor.Name

	return []*corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.harbor.NormalizeComponentName(containerregistryv1alpha1.RegistryName),
				Namespace: r.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.RegistryName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"REGISTRY_HTTP_SECRET": password.MustGenerate(keyLength, 5, 5, false, true),
			},
		},
	}
}

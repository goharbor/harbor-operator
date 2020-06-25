package core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	ConfigName = "app.conf"
)

func (r *Reconciler) GetConfigMap(ctx context.Context, core *goharborv1alpha2.Core) (*corev1.ConfigMap, error) {
	content, err := r.GetTemplatedConfig(ctx, ConfigTemplateKey, core)
	if err != nil {
		return nil, err
	}

	name := r.NormalizeName(ctx, core.GetName())
	namespace := core.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		BinaryData: map[string][]byte{
			ConfigName: content,
		},
	}, nil
}

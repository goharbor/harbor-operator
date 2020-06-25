package chartmuseum

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	ConfigName = "config.yaml"
)

// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/chartserver/env.jinja

func (r *Reconciler) GetConfigMap(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*corev1.ConfigMap, error) {
	content, err := r.GetTemplatedConfig(ctx, ConfigTemplateKey, chartMuseum)
	if err != nil {
		return nil, err
	}

	name := r.NormalizeName(ctx, chartMuseum.GetName())
	namespace := chartMuseum.GetNamespace()

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

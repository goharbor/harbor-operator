package notaryserver

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConfigName = "server.json"
)

func (r *Reconciler) GetConfigMap(ctx context.Context, notary *goharborv1alpha2.NotaryServer) (*corev1.ConfigMap, error) {
	content, err := r.GetTemplatedConfig(ctx, ConfigTemplateKey, notary)
	if err != nil {
		return nil, err
	}

	name := r.NormalizeName(ctx, notary.GetName())
	namespace := notary.GetNamespace()

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

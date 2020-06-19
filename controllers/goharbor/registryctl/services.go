package registryctl

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	PublicPort = 80
)

func (r *Reconciler) GetService(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(apiPort),
					Port:       PublicPort,
				},
			},
			Selector: map[string]string{
				"registryctl.goharbor.io/name":      name,
				"registryctl.goharbor.io/namespace": namespace,
			},
		},
	}, nil
}

package registry

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

func (r *Reconciler) GetService(ctx context.Context, registry *goharborv1alpha2.Registry) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, registry.GetName())
	namespace := registry.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       goharborv1alpha2.RegistryAPIPortName,
				Port:       registry.Spec.HTTP.TLS.GetInternalPort(),
				TargetPort: intstr.FromString(goharborv1alpha2.RegistryAPIPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

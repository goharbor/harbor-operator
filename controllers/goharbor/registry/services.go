package registry

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, registry *goharborv1.Registry) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, registry.GetName())
	namespace := registry.GetNamespace()

	ports := []corev1.ServicePort{{
		Name:       harbormetav1.RegistryAPIPortName,
		Port:       registry.Spec.HTTP.TLS.GetInternalPort(),
		TargetPort: intstr.FromString(harbormetav1.RegistryAPIPortName),
		Protocol:   corev1.ProtocolTCP,
	}}

	var annotations map[string]string

	if registry.Spec.HTTP.Debug != nil && registry.Spec.HTTP.Debug.Prometheus.Enabled {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.RegistryMetricsPortName,
			Port:       registry.Spec.HTTP.Debug.Port,
			TargetPort: intstr.FromString(harbormetav1.RegistryMetricsPortName),
			Protocol:   corev1.ProtocolTCP,
		})

		annotations = harbormetav1.AddPrometheusAnnotations(
			annotations,
			registry.Spec.HTTP.Debug.Port,
			registry.Spec.HTTP.Debug.Prometheus.Path,
		)
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

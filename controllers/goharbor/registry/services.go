package registry

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
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

func (r *Reconciler) GetCtlService(ctx context.Context, registryCtl *goharborv1.RegistryController) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	var ports []corev1.ServicePort

	if registryCtl.Spec.TLS.Enabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.RegistryControllerHTTPSPortName,
			Port:       harbormetav1.HTTPSPort,
			TargetPort: intstr.FromString(harbormetav1.RegistryControllerHTTPSPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	} else {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.RegistryControllerHTTPPortName,
			Port:       harbormetav1.HTTPPort,
			TargetPort: intstr.FromString(harbormetav1.RegistryControllerHTTPPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.NormalizeName(registryCtl.GetName(), RegistryCtlName),
			Namespace: namespace,
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

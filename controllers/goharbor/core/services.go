package core

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, core *goharborv1.Core) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, core.GetName())
	namespace := core.GetNamespace()
	annotations := map[string]string{}

	var ports []corev1.ServicePort

	if core.Spec.Components.TLS.Enabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.CoreHTTPSPortName,
			Port:       harbormetav1.HTTPSPort,
			TargetPort: intstr.FromString(harbormetav1.CoreHTTPSPortName),
			Protocol:   corev1.ProtocolTCP,
		})

		if v, ok := core.Annotations[harbormetav1.IngressControllerAnnotationName]; ok && v == string(harbormetav1.IngressControllerContour) {
			annotations["projectcontour.io/upstream-protocol.tls"] = harbormetav1.PortalHTTPSPortName
		}
	} else {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.CoreHTTPPortName,
			Port:       harbormetav1.HTTPPort,
			TargetPort: intstr.FromString(harbormetav1.CoreHTTPPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if core.Spec.Metrics.IsEnabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.CoreMetricsPortName,
			Port:       core.Spec.Metrics.Port,
			TargetPort: intstr.FromString(harbormetav1.CoreMetricsPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: core.Spec.Metrics.AddPrometheusAnnotations(annotations),
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

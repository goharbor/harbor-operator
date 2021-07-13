package jobservice

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, jobservice *goharborv1.JobService) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, jobservice.GetName())
	namespace := jobservice.GetNamespace()

	var ports []corev1.ServicePort

	if jobservice.Spec.TLS.Enabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.JobServiceHTTPSPortName,
			Port:       harbormetav1.HTTPSPort,
			TargetPort: intstr.FromString(harbormetav1.JobServiceHTTPSPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	} else {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.JobServiceHTTPPortName,
			Port:       harbormetav1.HTTPPort,
			TargetPort: intstr.FromString(harbormetav1.JobServiceHTTPPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if jobservice.Spec.Metrics.IsEnabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.JobServiceMetricsPortName,
			Port:       jobservice.Spec.Metrics.Port,
			TargetPort: intstr.FromString(harbormetav1.JobServiceMetricsPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: jobservice.Spec.Metrics.AddPrometheusAnnotations(nil),
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

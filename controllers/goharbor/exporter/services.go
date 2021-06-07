package exporter

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, exporter *goharborv1.Exporter) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, exporter.GetName())
	namespace := exporter.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: harbormetav1.AddPrometheusAnnotations(nil, exporter.Spec.Port, exporter.Spec.Path),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.ExporterMetricsPortName,
				Port:       exporter.Spec.Port,
				TargetPort: intstr.FromString(harbormetav1.ExporterMetricsPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

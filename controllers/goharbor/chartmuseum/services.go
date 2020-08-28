package chartmuseum

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, chartMuseum.GetName())
	namespace := chartMuseum.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.ChartMuseumHTTPPortName,
				Port:       harbormetav1.HTTPPort,
				TargetPort: intstr.FromString(harbormetav1.ChartMuseumHTTPPortName),
			}, {
				Name:       harbormetav1.ChartMuseumHTTPSPortName,
				Port:       harbormetav1.HTTPSPort,
				TargetPort: intstr.FromString(harbormetav1.ChartMuseumHTTPSPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

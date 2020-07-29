package chartmuseum

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
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
				Name:       goharborv1alpha2.ChartMuseumHTTPPortName,
				Port:       goharborv1alpha2.HTTPPort,
				TargetPort: intstr.FromString(goharborv1alpha2.ChartMuseumHTTPPortName),
			}, {
				Name:       goharborv1alpha2.ChartMuseumHTTPSPortName,
				Port:       goharborv1alpha2.HTTPSPort,
				TargetPort: intstr.FromString(goharborv1alpha2.ChartMuseumHTTPSPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

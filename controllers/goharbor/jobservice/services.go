package jobservice

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, jobservice.GetName())
	namespace := jobservice.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.JobServiceHTTPPortName,
				Port:       harbormetav1.HTTPPort,
				TargetPort: intstr.FromString(harbormetav1.JobServiceHTTPPortName),
			}, {
				Name:       harbormetav1.JobServiceHTTPSPortName,
				Port:       harbormetav1.HTTPSPort,
				TargetPort: intstr.FromString(harbormetav1.JobServiceHTTPSPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

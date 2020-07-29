package jobservice

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
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
				Name:       goharborv1alpha2.JobServiceHTTPPortName,
				Port:       goharborv1alpha2.HTTPPort,
				TargetPort: intstr.FromString(goharborv1alpha2.JobServiceHTTPPortName),
			}, {
				Name:       goharborv1alpha2.JobServiceHTTPSPortName,
				Port:       goharborv1alpha2.HTTPSPort,
				TargetPort: intstr.FromString(goharborv1alpha2.JobServiceHTTPSPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

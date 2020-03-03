package chartmuseum

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	PublicPort = 80
)

func (r *Reconciler) GetService(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-chartmuseum", chartMuseum.GetName()),
			Namespace: chartMuseum.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       PublicPort,
					TargetPort: intstr.FromInt(port),
				},
			},
			Selector: map[string]string{
				"chartmuseum-name":      chartMuseum.GetName(),
				"chartmuseum-namespace": chartMuseum.GetNamespace(),
			},
		},
	}, nil
}

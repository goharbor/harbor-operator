package registry

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

func (r *Reconciler) GetService(ctx context.Context, registry *goharborv1alpha2.Registry) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-registry", registry.GetName()),
			Namespace: registry.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "registry",
					TargetPort: intstr.FromInt(apiPort),
					Port:       PublicPort,
				}, {
					Name: "registry-debug",
					Port: metricsPort,
				},
			},
			Selector: map[string]string{
				"registry-name":      registry.GetName(),
				"registry-namespace": registry.GetNamespace(),
			},
		},
	}, nil
}

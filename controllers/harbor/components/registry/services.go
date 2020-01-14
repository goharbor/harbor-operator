package registry

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	PublicPort = 80
)

func (r *Registry) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := r.harbor.Name

	return []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.harbor.NormalizeComponentName(containerregistryv1alpha1.RegistryName),
				Namespace: r.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.RegistryName,
					"harbor":   harborName,
					"operator": operatorName,
				},
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
					}, {
						Name: "controller",
						Port: ctlAPIPort,
					},
				},
				Selector: map[string]string{
					"app":      containerregistryv1alpha1.RegistryName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}

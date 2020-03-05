package chartmuseum

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
	"github.com/goharbor/harbor-core-operator/pkg/factories/application"
)

const (
	PublicPort = 80
)

func (c *ChartMuseum) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port:       PublicPort,
						TargetPort: intstr.FromInt(port),
					},
				},
				Selector: map[string]string{
					"app":      containerregistryv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}

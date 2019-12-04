package portal

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

func (p *Portal) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := p.harbor.Name

	return []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.harbor.NormalizeComponentName(containerregistryv1alpha1.PortalName),
				Namespace: p.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.PortalName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.FromInt(8080),
					},
				},
				Selector: map[string]string{
					"app":      containerregistryv1alpha1.PortalName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}

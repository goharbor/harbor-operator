package portal

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

const (
	PublicPort = 80
)

func (p *Portal) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := p.harbor.Name

	return []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.harbor.NormalizeComponentName(goharborv1alpha1.PortalName),
				Namespace: p.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.PortalName,
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
					"app":      goharborv1alpha1.PortalName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}

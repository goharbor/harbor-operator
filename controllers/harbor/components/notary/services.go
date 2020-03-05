package notary

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/goharbor/harbor-core-operator/pkg/factories/application"
)

const (
	PublicPort = 80
)

func (n *Notary) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := n.harbor.Name

	return []*corev1.Service{
		{
			// https://github.com/goharbor/harbor-helm/blob/master/templates/notary/notary-svc.yaml
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(NotaryServerName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      NotaryServerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       NotaryServerName,
						Port:       PublicPort,
						TargetPort: intstr.FromInt(notaryServerPort),
					},
				},
				Selector: map[string]string{
					"app":      NotaryServerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
		{
			// https://github.com/goharbor/harbor-helm/blob/master/templates/notary/notary-svc.yaml
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(NotarySignerName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      NotarySignerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       NotarySignerName,
						Port:       PublicPort,
						TargetPort: intstr.FromInt(notarySignerPort),
					},
				},
				Selector: map[string]string{
					"app":      NotarySignerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}

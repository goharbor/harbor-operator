package clair

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

const (
	PublicPort        = 80
	AdapterPublicPort = 8080
)

func (c *Clair) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.Service{
		{
			// https://github.com/goharbor/harbor-helm/blob/master/templates/clair/clair-svc.yaml
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(goharborv1alpha1.ClairName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.ClairName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "api",
						Port:       PublicPort,
						TargetPort: intstr.FromInt(apiPort),
					}, {
						Name: "healthcheck",
						Port: healthPort,
					}, {
						Name:       "adapter",
						Port:       AdapterPublicPort,
						TargetPort: intstr.FromInt(adapterPort),
					},
				},
				Selector: map[string]string{
					"app":      goharborv1alpha1.ClairName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}

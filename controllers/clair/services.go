package clair

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	PublicPort        = 80
	AdapterPublicPort = 8080
)

func (r *Reconciler) GetService(ctx context.Context, clair *goharborv1alpha2.Clair) (*corev1.Service, error) {
	return &corev1.Service{
		// https://github.com/goharbor/harbor-helm/blob/master/templates/clair/clair-svc.yaml
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-clair", clair.GetName()),
			Namespace: clair.GetNamespace(),
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
				"clair-name":      clair.GetName(),
				"clair-namespace": clair.GetNamespace(),
			},
		},
	}, nil
}

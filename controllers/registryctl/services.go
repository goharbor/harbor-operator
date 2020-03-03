package registryctl

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

func (r *Reconciler) GetService(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-registryctl", registryCtl.GetName()),
			Namespace: registryCtl.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(ctlAPIPort),
					Port:       PublicPort,
				},
			},
			Selector: map[string]string{
				"registryctl-name":      registryCtl.GetName(),
				"registryctl-namespace": registryCtl.GetNamespace(),
			},
		},
	}, nil
}

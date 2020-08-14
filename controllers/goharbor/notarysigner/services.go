package notarysigner

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	PublicPort = 80
)

func (r *Reconciler) GetService(ctx context.Context, notary *goharborv1alpha2.NotarySigner) (*corev1.Service, error) {
	return &corev1.Service{
		// https://github.com/goharbor/harbor-helm/blob/master/templates/notary/notary-svc.yaml
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-notarysigner", notary.GetName()),
			Namespace: notary.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       notary.Name,
					Port:       PublicPort,
					TargetPort: intstr.FromInt(notarySignerPort),
				},
			},
			Selector: map[string]string{
				"notarysigner-name":      notary.GetName(),
				"notarysigner-namespace": notary.GetNamespace(),
			},
		},
	}, nil
}

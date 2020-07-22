package trivy

import (
	"context"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	PublicPort = 80
)

func (r *Reconciler) AddService(ctx context.Context, trivy *goharborv1alpha2.Trivy) error {
	// Forge the service resource
	service, err := r.GetService(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	// Add service to reconciler controller
	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot manage service %s", service.GetName())
	}

	return nil
}

func (r *Reconciler) GetService(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       PublicPort,
					TargetPort: intstr.FromInt(port),
				},
			},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

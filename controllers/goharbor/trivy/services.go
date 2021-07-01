package trivy

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) AddService(ctx context.Context, trivy *goharborv1.Trivy) error {
	service, err := r.GetService(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "get")
	}

	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrap(err, "add")
	}

	return nil
}

func (r *Reconciler) GetService(ctx context.Context, trivy *goharborv1.Trivy) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	var ports []corev1.ServicePort

	if trivy.Spec.Server.TLS.Enabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.TrivyHTTPSPortName,
			Port:       harbormetav1.HTTPSPort,
			TargetPort: intstr.FromString(harbormetav1.TrivyHTTPSPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	} else {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.TrivyHTTPPortName,
			Port:       harbormetav1.HTTPPort,
			TargetPort: intstr.FromString(harbormetav1.TrivyHTTPPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

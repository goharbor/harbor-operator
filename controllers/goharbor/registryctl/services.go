package registryctl

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PublicPort = 80
)

func (r *Reconciler) GetService(ctx context.Context, registryCtl *goharborv1.RegistryController) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	var ports []corev1.ServicePort

	if registryCtl.Spec.TLS.Enabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.RegistryControllerHTTPSPortName,
			Port:       harbormetav1.HTTPSPort,
			TargetPort: intstr.FromString(harbormetav1.RegistryControllerHTTPSPortName),
			Protocol:   corev1.ProtocolTCP,
		})
	} else {
		ports = append(ports, corev1.ServicePort{
			Name:       harbormetav1.RegistryControllerHTTPPortName,
			Port:       harbormetav1.HTTPPort,
			TargetPort: intstr.FromString(harbormetav1.RegistryControllerHTTPPortName),
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

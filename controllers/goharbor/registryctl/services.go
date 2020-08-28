package registryctl

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PublicPort = 80
)

func (r *Reconciler) GetService(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.RegistryControllerHTTPPortName,
				Port:       harbormetav1.HTTPPort,
				TargetPort: intstr.FromString(harbormetav1.RegistryControllerHTTPPortName),
			}, {
				Name:       harbormetav1.RegistryControllerHTTPSPortName,
				Port:       harbormetav1.HTTPSPort,
				TargetPort: intstr.FromString(harbormetav1.RegistryControllerHTTPSPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

package notarysigner

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

const (
	PublicPort = 80
)

func (r *Reconciler) GetService(ctx context.Context, notary *goharborv1alpha2.NotarySigner) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, notary.GetName())
	namespace := notary.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.NotarySignerAPIPortName,
				Port:       goharborv1alpha2.NotarySignerAPIPort,
				TargetPort: intstr.FromString(harbormetav1.NotarySignerAPIPortName),
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

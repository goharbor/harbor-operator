package notaryserver

import (
	"context"
	"fmt"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PublicPort = 443
)

func (r *Reconciler) GetService(ctx context.Context, notary *goharborv1alpha2.NotaryServer) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, notary.GetName())
	namespace := notary.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.NotaryServerAPIPortName,
				Port:       notary.Spec.TLS.GetInternalPort(),
				TargetPort: intstr.FromString(harbormetav1.NotaryServerAPIPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      name,
				r.Label("namespace"): namespace,
			},
		},
	}, nil
}

func GetLBService(svc *corev1.Service) *corev1.Service {
	svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	svc.Name = fmt.Sprintf("%s-lb", svc.Name)

	return svc
}

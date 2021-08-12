package notaryserver

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) GetService(ctx context.Context, notary *goharborv1.NotaryServer) (*corev1.Service, error) {
	name := r.NormalizeName(ctx, notary.GetName())
	namespace := notary.GetNamespace()
	annotations := map[string]string{}

	if v, ok := notary.Annotations[harbormetav1.IngressControllerAnnotationName]; ok && v == string(harbormetav1.IngressControllerContour) {
		annotations["projectcontour.io/upstream-protocol.tls"] = harbormetav1.NotaryServerAPIPortName
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
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

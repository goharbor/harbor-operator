package harbor

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor-operator/controllers"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type CoreLB graph.Resource

func (r *Reconciler) AddCoreLB(ctx context.Context, harbor *goharborv1alpha2.Harbor, core Core, portal Portal) (CoreLB, error) {

	ingressRes, err := r.Controller.AddServiceToManage(ctx, r.GetCoreLB(ctx, harbor), core, portal)

	return CoreLB(ingressRes), errors.Wrap(err, "cannot add core lb service")
}

func (r *Reconciler) GetCoreLB(ctx context.Context, harbor *goharborv1alpha2.Harbor) *corev1.Service {
	if harbor.Spec.Expose.Core.LoadBalancer == nil {
		return nil
	}

	core := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String())

	name := fmt.Sprintf("%s-lb", core)
	namespace := harbor.GetNamespace()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       harbormetav1.CoreHTTPPortName,
				Port:       harbormetav1.HTTPPort,
				TargetPort: intstr.FromString(harbormetav1.CoreHTTPPortName),
				Protocol:   corev1.ProtocolTCP,
			}, {
				Name:       harbormetav1.CoreHTTPSPortName,
				Port:       harbormetav1.HTTPSPort,
				TargetPort: intstr.FromString(harbormetav1.CoreHTTPSPortName),
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				r.Label("name"):      core,
				r.Label("namespace"): namespace,
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

}

type NotaryLB graph.Resource

func (r *Reconciler) NotaryLB(ctx context.Context, harbor *goharborv1alpha2.Harbor, core Core, portal Portal) (NotaryLB, error) {

	ingressRes, err := r.Controller.AddServiceToManage(ctx, r.GetCoreLB(ctx, harbor), core, portal)

	return NotaryLB(ingressRes), errors.Wrap(err, "cannot add core ingress")
}

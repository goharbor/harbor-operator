package mutation

import (
	"context"

	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

type MutateIngress func(context.Context, *netv1.Ingress, *netv1.Ingress) controllerutil.MutateFn

func NewIngress(ingress *netv1.Ingress, mutate MutateIngress) resources.Mutable {
	return func(ctx context.Context, ingressResource, ingressResult runtime.Object) controllerutil.MutateFn {
		result := ingressResult.(*netv1.Ingress)
		previous := ingressResource.(*netv1.Ingress)

		mutate := mutate(ctx, previous, result)

		return func() error {
			previous.DeepCopyInto(result)

			return mutate()
		}
	}
}

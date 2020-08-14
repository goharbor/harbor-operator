package mutation

import (
	"context"

	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func NewIngress(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, ingressResource, ingressResult runtime.Object) controllerutil.MutateFn {
		result := ingressResult.(*netv1.Ingress)
		desired := ingressResource.(*netv1.Ingress)

		mutate := mutate(ctx, desired, result)

		return func() error {
			desired.Spec.DeepCopyInto(&result.Spec)

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}

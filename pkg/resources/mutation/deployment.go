package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewDeployment(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, deploymentResource, deploymentResult runtime.Object) controllerutil.MutateFn {
		result := deploymentResult.(*appsv1.Deployment)
		desired := deploymentResource.(*appsv1.Deployment)

		mutate := mutate(ctx, desired, result)

		return func() error {
			desired.Spec.DeepCopyInto(&result.Spec)

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}

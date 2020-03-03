package mutation

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

type MutateDeployment func(context.Context, *appsv1.Deployment, *appsv1.Deployment) controllerutil.MutateFn

func NewDeployment(deployment *appsv1.Deployment, mutate MutateDeployment) resources.Mutable {
	return func(ctx context.Context, deploymentResource, deploymentResult runtime.Object) controllerutil.MutateFn {
		result := deploymentResult.(*appsv1.Deployment)
		previous := deploymentResource.(*appsv1.Deployment)

		mutate := mutate(ctx, previous, result)

		return func() error {
			previous.DeepCopyInto(result)

			return mutate()
		}
	}
}

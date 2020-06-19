package mutation

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func NewConfigMap(mutate resources.Mutable) resources.Mutable {
	return func(ctx context.Context, configResource, configResult runtime.Object) controllerutil.MutateFn {
		result := configResult.(*corev1.ConfigMap)
		desired := configResource.(*corev1.ConfigMap)

		mutate := mutate(ctx, desired, result)

		return func() error {
			result.BinaryData = desired.BinaryData
			result.Data = desired.Data
			result.Immutable = desired.Immutable

			return mutate()
		}
	}
}

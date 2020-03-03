package mutation

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

type MutateConfigMap func(context.Context, *corev1.ConfigMap, *corev1.ConfigMap) controllerutil.MutateFn

func NewConfigMap(configMap *corev1.ConfigMap, mutate MutateConfigMap) resources.Mutable {
	return func(ctx context.Context, configResource, configResult runtime.Object) controllerutil.MutateFn {
		result := configResult.(*corev1.ConfigMap)
		previous := configResource.(*corev1.ConfigMap)

		mutate := mutate(ctx, previous, result)

		return func() error {
			previous.DeepCopyInto(result)

			return mutate()
		}
	}
}

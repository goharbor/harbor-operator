package mutation

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

type MutateSecret func(context.Context, *corev1.Secret, *corev1.Secret) controllerutil.MutateFn

func NewSecret(mutate MutateSecret) resources.Mutable {
	return func(ctx context.Context, secretResource, secretResult runtime.Object) controllerutil.MutateFn {
		result := secretResult.(*corev1.Secret)
		previous := secretResource.(*corev1.Secret)

		mutate := mutate(ctx, previous, result)

		return func() error {
			// Most of password are generated
			// Do not override existing secrets
			// To update secrets value, we should rename the key or
			//  delete it before recreating it.
			for key := range result.Data {
				_, okString := previous.StringData[key]
				_, okBytes := previous.Data[key]

				if !okString && !okBytes {
					delete(result.Data, key)
				}

				delete(previous.Data, key)
				delete(previous.StringData, key)
			}

			if result.Data == nil {
				result.Data = map[string][]byte{}
			}

			for name, value := range previous.Data {
				result.Data[name] = value
			}

			// StringData is write only, it overrides Data
			// so we can compare with remote result
			result.StringData = previous.StringData

			return mutate()
		}
	}
}

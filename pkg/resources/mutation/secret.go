package mutation

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func NewSecret(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, secretResource, secretResult runtime.Object) controllerutil.MutateFn {
		result := secretResult.(*corev1.Secret)
		desired := secretResource.(*corev1.Secret)

		mutate := mutate(ctx, desired, result)

		return func() error {
			// Most of password are generated
			// Do not override existing secrets
			// To update secrets value, we should rename the key or
			//  delete it before recreating it.
			for key := range result.Data {
				_, okString := desired.StringData[key]
				_, okBytes := desired.Data[key]

				if !okString && !okBytes {
					delete(result.Data, key)
				}

				delete(desired.Data, key)
				delete(desired.StringData, key)
			}

			if result.Data == nil {
				result.Data = map[string][]byte{}
			}

			for name, value := range desired.Data {
				result.Data[name] = value
			}

			// StringData is write only, it overrides Data
			// so we can compare with remote result
			result.StringData = desired.StringData

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}

package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewSecret(mutate resources.Mutable, override, remove bool) (result resources.Mutable) {
	result = func(ctx context.Context, secretResource, secretResult runtime.Object) controllerutil.MutateFn {
		result := secretResult.(*corev1.Secret)
		desired := secretResource.(*corev1.Secret)

		mutate := mutate(ctx, desired, result)

		return func() error {
			for key := range result.Data {
				_, okString := desired.StringData[key]
				_, okBytes := desired.Data[key]

				if remove && !okString && !okBytes {
					delete(result.Data, key)
				}

				if !override {
					delete(desired.Data, key)
					delete(desired.StringData, key)
				}
			}

			if result.Data == nil {
				result.Data = map[string][]byte{}
			}

			for name, value := range desired.Data {
				result.Data[name] = value
			}

			if result.StringData == nil {
				result.StringData = map[string]string{}
			}

			for name, value := range desired.StringData {
				result.StringData[name] = value
			}

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}

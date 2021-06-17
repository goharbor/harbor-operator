package template

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSecretDataFunc(ctx context.Context, c client.Client, namespace string, ignoreNotFound bool) interface{} {
	return GetK8SNamespacedDataFunc(ctx, c, namespace, &corev1.Secret{}, func(ctx context.Context, o client.Object) (map[string]interface{}, error) {
		secret := o.(*corev1.Secret)
		result := make(map[string]interface{}, len(secret.Data))

		for k, d := range secret.Data {
			result[k] = string(d)
		}

		return result, nil
	}, ignoreNotFound)
}

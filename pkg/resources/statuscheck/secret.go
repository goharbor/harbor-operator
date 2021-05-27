package statuscheck

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SecretCheck(ctx context.Context, object client.Object, keys ...string) (bool, error) {
	secret := object.(*corev1.Secret)

	for _, key := range keys {
		if data, ok := secret.Data[key]; !ok || len(data) == 0 {
			return false, nil
		}
	}

	return true, nil
}

func TLSSecretCheck(ctx context.Context, object client.Object) (bool, error) {
	return SecretCheck(ctx, object, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.ServiceAccountRootCAKey)
}

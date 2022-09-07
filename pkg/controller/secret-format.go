package controller

import (
	"context"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errSecretInvalidTyped = errors.New("unexpected secret type")

func (c *Controller) EnsureSecretType(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return errors.Errorf("unsupported resource type %+v", node)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "checkSecretType")
	defer span.Finish()

	gvk := c.AddGVKToSpan(ctx, span, res.resource)

	objectKey := client.ObjectKeyFromObject(res.resource)

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	secret := &corev1.Secret{}

	err := c.Client.Get(ctx, objectKey, secret)
	if err != nil {
		// TODO Check if the error is a temporary error or a unrecoverrable one
		return errors.Wrapf(err, "cannot get %s %s/%s", gvk, res.resource.GetNamespace(), res.resource.GetName())
	}

	expectedSecretType := res.resource.(*corev1.Secret).Type

	switch secret.Type { //nolint:exhaustive
	default:
		return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
	case res.resource.(*corev1.Secret).Type:
		return nil
	case corev1.SecretTypeOpaque:
		switch expectedSecretType { //nolint:exhaustive
		default:
			return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
		case harbormetav1.SecretTypeRedis:
			if isOpaqueRedisSecret(secret) {
				return nil
			}

			return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
		case harbormetav1.SecretTypePostgresql:
			if isOpaquePostgresqlSecret(secret) {
				return nil
			}

			return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
		}
	}
}

func isOpaqueRedisSecret(secret *corev1.Secret) bool {
	if len(secret.Data) > 0 {
		if _, ok := secret.Data[harbormetav1.RedisPasswordKey]; ok {
			return true
		}
	}

	if len(secret.StringData) > 0 {
		if _, ok := secret.StringData[harbormetav1.RedisPasswordKey]; ok {
			return true
		}
	}

	return false
}

func isOpaquePostgresqlSecret(secret *corev1.Secret) bool {
	if len(secret.Data) > 0 {
		if _, ok := secret.Data[harbormetav1.PostgresqlPasswordKey]; ok {
			return true
		}
	}

	if len(secret.StringData) > 0 {
		if _, ok := secret.StringData[harbormetav1.PostgresqlPasswordKey]; ok {
			return true
		}
	}

	return false
}

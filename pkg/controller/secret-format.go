package controller

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
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

	objectKey, err := client.ObjectKeyFromObject(res.resource)
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	secret := &corev1.Secret{}

	err = c.Client.Get(ctx, objectKey, secret)
	if err != nil {
		// TODO Check if the error is a temporary error or a unrecoverrable one
		return errors.Wrapf(err, "cannot get %s %s/%s", gvk, res.resource.GetNamespace(), res.resource.GetName())
	}

	expectedSecretType := res.resource.(*corev1.Secret).Type

	switch secret.Type {
	default:
		return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
	case res.resource.(*corev1.Secret).Type:
		return nil
	case corev1.SecretTypeOpaque:
		switch expectedSecretType {
		default:
			return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
		case goharborv1alpha2.SecretTypeRedis:
			if isOpaqueRedisSecret(secret) {
				return nil
			}

			return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
		case goharborv1alpha2.SecretTypePostgresql:
			if isOpaquePostgresqlSecret(secret) {
				return nil
			}

			return errors.Wrapf(errSecretInvalidTyped, "got %s expected %s", secret.Type, expectedSecretType)
		}
	}
}

func isOpaqueRedisSecret(secret *corev1.Secret) bool {
	if len(secret.Data) > 0 {
		if _, ok := secret.Data[goharborv1alpha2.RedisPasswordKey]; ok {
			return true
		}
	}

	if len(secret.StringData) > 0 {
		if _, ok := secret.StringData[goharborv1alpha2.RedisPasswordKey]; ok {
			return true
		}
	}

	return false
}

func isOpaquePostgresqlSecret(secret *corev1.Secret) bool {
	if len(secret.Data) > 0 {
		if _, ok := secret.Data[goharborv1alpha2.PostgresqlPasswordKey]; ok {
			return true
		}
	}

	if len(secret.StringData) > 0 {
		if _, ok := secret.StringData[goharborv1alpha2.PostgresqlPasswordKey]; ok {
			return true
		}
	}

	return false
}

package core

import (
	"context"
	"fmt"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	RedisDSNKey         = "_REDIS_URL"
	RegistryRedisDSNKey = "_REDIS_URL_REG"
)

func (r *Reconciler) GetSecret(ctx context.Context, core *goharborv1alpha2.Core) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, core.GetName())
	namespace := core.GetNamespace()

	var redisPassword string

	if core.Spec.Redis.PasswordRef != "" {
		var passwordSecret corev1.Secret

		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      core.Spec.Redis.PasswordRef,
		}, &passwordSecret)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get redis password")
		}

		password, ok := passwordSecret.Data[harbormetav1.RedisPasswordKey]
		if !ok {
			return nil, errors.Errorf("%s not found in secret %s", harbormetav1.RedisPasswordKey, core.Spec.Redis.PasswordRef)
		}

		redisPassword = string(password)
	}

	var registryPassword string

	if core.Spec.Components.Registry.Redis != nil && core.Spec.Components.Registry.Redis.PasswordRef != "" {
		var passwordSecret corev1.Secret

		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      core.Spec.Components.Registry.Redis.PasswordRef,
		}, &passwordSecret)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get registry redis password")
		}

		password, ok := passwordSecret.Data[harbormetav1.RedisPasswordKey]
		if !ok {
			return nil, errors.Errorf("%s not found in secret %s", harbormetav1.RedisPasswordKey, core.Spec.Components.Registry.Redis.PasswordRef)
		}

		registryPassword = string(password)
	}

	registryCacheDSN := core.Spec.Components.Registry.Redis.GetDSNStringWithRawPassword(registryPassword)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			RedisDSNKey:         fmt.Sprintf("%s:%d,100,%s,%d,%.0f", core.Spec.Redis.Host, core.Spec.Redis.Port, redisPassword, core.Spec.Redis.Database, core.Spec.Redis.IdleTimeout.Duration.Seconds()),
			RegistryRedisDSNKey: registryCacheDSN,
		},
	}, nil
}

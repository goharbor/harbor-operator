package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	RedisDSNKey         = "_REDIS_URL"
	RegistryRedisDSNKey = "_REDIS_URL_REG"
)

func (r *Reconciler) GetSecret(ctx context.Context, core *goharborv1alpha2.Core) (*corev1.Secret, error) { // nolint:funlen
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

		password, ok := passwordSecret.Data[goharborv1alpha2.RedisPasswordKey]
		if !ok {
			return nil, errors.Errorf("%s not found in secret %s", goharborv1alpha2.RedisPasswordKey, core.Spec.Redis.PasswordRef)
		}

		redisPassword = string(password)
	}

	redisDSN, err := core.Spec.Redis.GetDSN(redisPassword)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get redis DSN")
	}

	var registryPassword string

	if core.Spec.Components.Registry.Redis.PasswordRef != "" {
		var passwordSecret corev1.Secret

		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      core.Spec.Components.Registry.Redis.PasswordRef,
		}, &passwordSecret)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get registry redis password")
		}

		password, ok := passwordSecret.Data[goharborv1alpha2.RedisPasswordKey]
		if !ok {
			return nil, errors.Errorf("%s not found in secret %s", goharborv1alpha2.RedisPasswordKey, core.Spec.Components.Registry.Redis.PasswordRef)
		}

		registryPassword = string(password)
	}

	registryCacheDSN, err := core.Spec.Components.Registry.Redis.GetDSN(registryPassword)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get registry redis dsn")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			RedisDSNKey:         fmt.Sprintf("%s:%s,100,%s,%s,%.0f", redisDSN.Hostname(), redisDSN.Port(), redisPassword, strings.Trim(redisDSN.EscapedPath(), "/"), core.Spec.Redis.IdleTimeout.Seconds()),
			RegistryRedisDSNKey: registryCacheDSN.String(),
		},
	}, nil
}

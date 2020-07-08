package trivy

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

func (r *Reconciler) GetSecret(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()
	var redisPassword string

	if trivy.Spec.Cache.Redis.PasswordRef != "" {
		var passwordSecret corev1.Secret

		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      trivy.Spec.Cache.Redis.PasswordRef,
		}, &passwordSecret)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get redis password")
		}

		password, ok := passwordSecret.Data[goharborv1alpha2.RedisPasswordKey]
		if !ok {
			return nil, errors.Errorf("%s not found in secret %s", goharborv1alpha2.RedisPasswordKey, trivy.Spec.Cache.Redis.PasswordRef)
		}

		redisPassword = string(password)
	}

	redisDSN, err := trivy.Spec.Cache.Redis.GetDSN(redisPassword)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get redis DSN")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"SCANNER_REDIS_URL": redisDSN.String(),
		},
	}, nil
}

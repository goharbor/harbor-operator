package trivy

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/pkg/errors"
)

func (r *Reconciler) AddSecret(ctx context.Context, trivy *goharborv1alpha2.Trivy) error {
	// Forge the secret resource
	secret, err := r.GetSecret(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot get secret")
	}

	// Add secret to reconciler controller
	_, err = r.Controller.AddSecretToManage(ctx, secret)
	if err != nil {
		return errors.Wrapf(err, "cannot manage secret %s", secret.GetName())
	}

	return nil
}

func (r *Reconciler) GetSecret(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*corev1.Secret, error) {
	var redisPassword string

	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	if trivy.Spec.Cache.Redis.PasswordRef != "" {
		var passwordSecret corev1.Secret

		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      trivy.Spec.Cache.Redis.PasswordRef,
		}, &passwordSecret)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get redis password")
		}

		password, ok := passwordSecret.Data[harbormetav1.RedisPasswordKey]
		if !ok {
			return nil, errors.Errorf("%s not found in secret %s", harbormetav1.RedisPasswordKey, trivy.Spec.Cache.Redis.PasswordRef)
		}

		redisPassword = string(password)
	}

	redisDSN := trivy.Spec.Cache.Redis.GetDSN(redisPassword)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"SCANNER_REDIS_URL":           redisDSN.String(),
			"SCANNER_JOB_QUEUE_REDIS_URL": redisDSN.String(),
		},
	}, nil
}

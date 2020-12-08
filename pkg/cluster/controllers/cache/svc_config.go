package cache

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ lcm.SvcConfigGetter = &RedisConfigGetter{}

// NewRedisConfigGetter constructor for RedisConfigGetter.
func NewRedisConfigGetter() *RedisConfigGetter {
	return &RedisConfigGetter{}
}

// RedisConfigGetter implements SvcConfigGetter.
type RedisConfigGetter struct {
	ctx     context.Context
	client  client.Client
	cluster *v1alpha2.HarborCluster
}

// WithCtx binds ctx.
func (r *RedisConfigGetter) WithCtx(ctx context.Context) lcm.SvcConfigGetter {
	r.ctx = ctx

	return r
}

// UseClient binds client.
func (r *RedisConfigGetter) UseClient(client client.Client) lcm.SvcConfigGetter {
	r.client = client

	return r
}

// FromCluster from cluster info.
func (r *RedisConfigGetter) FromCluster(cluster *v1alpha2.HarborCluster) lcm.SvcConfigGetter {
	r.cluster = cluster

	return r
}

// GetConfig gets service config.
func (r *RedisConfigGetter) GetConfig() (*lcm.ServiceConfig, []lcm.Option, error) {
	if r.cluster == nil {
		return nil, nil, fmt.Errorf("cluster can not be nil, call FromCluster before GetConfig")
	}

	spec := r.cluster.Spec.HarborComponentsSpec.Redis
	if spec == nil {
		return nil, nil, fmt.Errorf("cluster redis spec can not be nil")
	}
	// get out-cluster redis svc config
	svcConfig := &lcm.ServiceConfig{
		Endpoint: &lcm.Endpoint{
			Host: spec.Host,
			Port: uint(spec.Port),
		},
	}

	if spec.PasswordRef != "" {
		secretName := spec.PasswordRef
		secretNamespace := r.cluster.Namespace
		secret := &corev1.Secret{}
		// get secret
		err := r.client.Get(r.ctx, types.NamespacedName{Namespace: secretNamespace, Name: secretName}, secret)
		if err != nil {
			return nil, nil, fmt.Errorf("get secret %s/%s failed, error: %w", secretNamespace, secretName, err)
		}
		// retrieve password
		password, ok := secret.Data[harbormetav1.RedisPasswordKey]
		if !ok {
			return nil, nil, fmt.Errorf("%s not found in secret %s/%s", harbormetav1.RedisPasswordKey, secretNamespace, secretName)
		}
		// add Credentials
		svcConfig.Credentials = &lcm.Credentials{AccessSecret: string(password)}
	}

	return svcConfig, nil, nil
}

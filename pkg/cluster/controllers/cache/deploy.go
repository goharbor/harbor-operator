package cache

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Deploy will deploy Redis sentinel cluster if that does not exist.
// It does:
// - check redis does exist
// - create any new RedisFailovers CRs
// - create redis password secret
// It does not:
// - perform any RedisFailovers downscale (left for downscale phase)
// - perform any RedisFailovers upscale (left for upscale phase)
// - perform any pod upgrade (left for rolling upgrade phase).
func (rc *RedisController) Deploy(ctx context.Context, cluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	crdClient := rc.DClient.DynamicClient(ctx, k8s.WithResource(redisFailoversGVR), k8s.WithNamespace(cluster.Namespace))

	expectCR, err := rc.ResourceManager.GetCacheCR(ctx, cluster)
	if err != nil {
		return cacheNotReadyStatus(ErrorGenerateRedisCr, err.Error()), err
	}

	if err := controllerutil.SetControllerReference(cluster, expectCR.(metav1.Object), rc.Scheme); err != nil {
		return cacheNotReadyStatus(ErrorSetOwnerReference, err.Error()), err
	}

	if err := rc.DeploySecret(ctx, cluster); err != nil {
		return cacheNotReadyStatus(ErrorCreateRedisSecret, err.Error()), err
	}

	rc.Log.Info("Creating Redis.", "namespace", cluster.Namespace, "name", cluster.Name)

	unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(expectCR)
	if err != nil {
		return cacheNotReadyStatus(ErrorDefaultUnstructuredConverter, err.Error()), err
	}

	if _, err = crdClient.Create(&unstructured.Unstructured{Object: unstructuredData}, metav1.CreateOptions{}); err != nil {
		return cacheNotReadyStatus(ErrorCreateRedisCr, err.Error()), err
	}

	rc.Log.Info("Redis has been created.", "namespace", cluster.Namespace, "name", cluster.Name)

	return cacheUnknownStatus(), nil
}

func (rc *RedisController) DeploySecret(ctx context.Context, cluster *goharborv1.HarborCluster) error {
	secret := &corev1.Secret{}

	sec := rc.ResourceManager.GetSecret()
	if err := controllerutil.SetControllerReference(cluster, sec, rc.Scheme); err != nil {
		return err
	}

	err := rc.Client.Get(ctx, types.NamespacedName{Name: sec.Name, Namespace: sec.Namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		rc.Log.Info("Creating Redis Password Secret", "namespace", sec.Namespace, "name", sec.Name)

		return rc.Client.Create(ctx, sec)
	}

	return err
}

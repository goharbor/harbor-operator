package cache

import (
	"fmt"

	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/lcm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Deploy reconcile will deploy Redis sentinel cluster if that does not exist.
// It does:
// - check redis does exist
// - create any new RedisFailovers CRs
// - create redis password secret
// It does not:
// - perform any RedisFailovers downscale (left for downscale phase)
// - perform any RedisFailovers upscale (left for upscale phase)
// - perform any pod upgrade (left for rolling upgrade phase)
func (redis *RedisReconciler) Deploy() (*lcm.CRStatus, error) {

	if redis.HarborCluster.Spec.Cache.Kind == "external" {
		return cacheUnknownStatus(), nil
	}

	var expectCR *unstructured.Unstructured

	crdClient := redis.DClient.WithResource(redisFailoversGVR).WithNamespace(redis.HarborCluster.Namespace)

	expectCR, err := redis.generateRedisCR()
	if err != nil {
		return cacheNotReadyStatus(GenerateRedisCrError, err.Error()), err
	}

	if err := controllerutil.SetControllerReference(redis.HarborCluster, expectCR, redis.Scheme); err != nil {
		return cacheNotReadyStatus(SetOwnerReferenceError, err.Error()), err
	}

	if err := redis.DeploySecret(); err != nil {
		return cacheNotReadyStatus(CreateRedisSecretError, err.Error()), err
	}

	if err := redis.DeployService(); err != nil {
		return cacheNotReadyStatus(CreateRedisServerServiceError, err.Error()), err
	}

	redis.Log.Info("Creating Redis.", "namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name)

	_, err = crdClient.Create(expectCR, metav1.CreateOptions{})
	if err != nil {
		return cacheNotReadyStatus(CreateRedisCrError, err.Error()), err
	}

	redis.Log.Info("Redis has been created.", "namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name)
	return cacheUnknownStatus(), nil
}

// DeploySecret deploy the Redis Password Secret
func (redis *RedisReconciler) DeploySecret() error {
	secret := &corev1.Secret{}
	sc := redis.generateRedisSecret()

	if err := controllerutil.SetControllerReference(redis.HarborCluster, sc, redis.Scheme); err != nil {
		return err
	}

	err := redis.Client.Get(types.NamespacedName{Name: redis.HarborCluster.Name, Namespace: redis.HarborCluster.Namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		redis.Log.Info("Creating Redis Password Secret", "namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name)
		return redis.Client.Create(sc)
	}

	return err
}

// DeploySecret deploy the Redis Password Secret
func (redis *RedisReconciler) DeployService() error {
	service := &corev1.Service{}
	name := fmt.Sprintf("%s-%s", "cluster", redis.GetHarborClusterName())
	svc := redis.generateService()

	if err := controllerutil.SetControllerReference(redis.HarborCluster, svc, redis.Scheme); err != nil {
		return err
	}

	err := redis.Client.Get(types.NamespacedName{Name: name, Namespace: redis.HarborCluster.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		redis.Log.Info("Creating Redis server service", "namespace", redis.HarborCluster.Namespace, "name", name)
		return redis.Client.Create(svc)
	}

	return err
}

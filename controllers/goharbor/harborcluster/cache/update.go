package cache

import (
	"fmt"

	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/k8s"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/lcm"
	"github.com/google/go-cmp/cmp"
	redisCli "github.com/spotahome/redis-operator/api/redisfailover/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	//appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RollingUpgrades reconcile will rolling upgrades Redis sentinel cluster if resource upscale.
// It does:
// - check resource
// - update RedisFailovers CR resource
func (redis *RedisReconciler) RollingUpgrades() (*lcm.CRStatus, error) {

	crdClient := redis.DClient.WithResource(redisFailoversGVR).WithNamespace(redis.HarborCluster.Namespace)
	if redis.ExpectCR == nil {
		return cacheUnknownStatus(), nil
	}

	var actualCR redisCli.RedisFailover
	var expectCR redisCli.RedisFailover

	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(redis.ActualCR.UnstructuredContent(), &actualCR); err != nil {
		return cacheNotReadyStatus(DefaultUnstructuredConverterError, err.Error()), err
	}

	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(redis.ExpectCR.UnstructuredContent(), &expectCR); err != nil {
		return cacheNotReadyStatus(DefaultUnstructuredConverterError, err.Error()), err
	}

	if !IsEqual(expectCR, actualCR) {
		msg := fmt.Sprintf(UpdateMessageRedisCluster, redis.HarborCluster.Name)
		redis.Recorder.Event(redis.HarborCluster, corev1.EventTypeNormal, RedisUpScaling, msg)

		redis.Log.Info(
			"Update Redis resource",
			"namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name,
		)

		if err := Update(crdClient, actualCR, expectCR); err != nil {
			return cacheNotReadyStatus(UpdateRedisCrError, err.Error()), err
		}
	}
	return cacheUnknownStatus(), nil
}

// isEqual check whether cache cr is equal expect.
func IsEqual(actualCR, expectCR redisCli.RedisFailover) bool {
	return cmp.Equal(expectCR.DeepCopy().Spec, actualCR.DeepCopy().Spec)
}

func Update(crdClient k8s.DClient, actualCR, expectCR redisCli.RedisFailover) error {
	expectCR.ObjectMeta.SetResourceVersion(actualCR.ObjectMeta.GetResourceVersion())

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&expectCR)
	if err != nil {
		return err
	}

	_, err = crdClient.Update(&unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

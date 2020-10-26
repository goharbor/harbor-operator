package cache

import (
	"fmt"

	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/common"

	redisCli "github.com/spotahome/redis-operator/api/redisfailover/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	redisFailoversGVR = redisCli.SchemeGroupVersion.WithResource(redisCli.RFNamePlural)
)

// generateRedisCR returns RedisFailovers CRs
func (redis *RedisReconciler) generateRedisCR() (*unstructured.Unstructured, error) {
	redisResource := redis.GetRedisResource()
	redisRep := redis.GetRedisServerReplica()
	sentinelRep := redis.GetRedisSentinelReplica()
	storageSize := redis.GetRedisStorageSize()

	conf := &redisCli.RedisFailover{
		TypeMeta: v1.TypeMeta{
			Kind:       "RedisFailover",
			APIVersion: "databases.spotahome.com/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      redis.HarborCluster.Name,
			Namespace: redis.HarborCluster.Namespace,
			Labels:    redis.Labels,
		},
		Spec: redisCli.RedisFailoverSpec{
			Redis: redisCli.RedisSettings{
				Replicas: redisRep,
				Resources: corev1.ResourceRequirements{
					Requests: redisResource,
					Limits:   redisResource,
				},
			},
			Sentinel: redisCli.SentinelSettings{
				Replicas: sentinelRep,
				Resources: corev1.ResourceRequirements{
					Requests: redisResource,
					Limits:   redisResource,
				},
			},
			Auth: redisCli.AuthSettings{SecretPath: redis.HarborCluster.Name},
		},
	}

	conf.Spec.Cache.Storage.PersistentVolumeClaim = redis.generateRedisStorage(storageSize, redis.HarborCluster.Name)

	mapResult, err := runtime.DefaultUnstructuredConverter.ToUnstructured(conf)
	if err != nil {
		return nil, err
	}
	data := unstructured.Unstructured{Object: mapResult}

	return &data, nil
}

//generateRedisSecret returns redis password secret
func (redis *RedisReconciler) generateRedisSecret() *corev1.Secret {

	passStr := common.RandomString(8, "a")

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.HarborCluster.Name,
			Namespace: redis.HarborCluster.Namespace,
			Labels:    redis.Labels,
		},
		StringData: map[string]string{
			"password": passStr,
		},
	}
}

func (redis *RedisReconciler) generateRedisStorage(size, name string) *corev1.PersistentVolumeClaim {
	storage, _ := resource.ParseQuantity(size)
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Labels: redis.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Selector: nil,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storage,
				},
			},
		},
	}
}

//generateRedisSecret returns redis password secret
func (redis *RedisReconciler) generateHarborCacheSecret(component, secretName, url, namespace string) *corev1.Secret {

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: redis.HarborCluster.Namespace,
		},
		StringData: map[string]string{
			"url":       url,
			"namespace": namespace,
		},
	}
}

func (redis *RedisReconciler) generateService() *corev1.Service {
	name := fmt.Sprintf("%s-%s", "cluster", redis.GetHarborClusterName())
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: redis.GetHarborClusterNamespace(),
			Labels:    redis.Labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app.kubernetes.io/component": "redis",
				"app.kubernetes.io/name":      redis.GetHarborClusterName(),
				"app.kubernetes.io/part-of":   "redis-failover",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
					Protocol:   corev1.ProtocolTCP,
					Name:       "redis",
				},
			},
		},
	}
}

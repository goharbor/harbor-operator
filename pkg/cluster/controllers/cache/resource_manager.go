package cache

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceManager defines the common interface of resources.
type ResourceManager interface {
	ResourceGetter
}

// ResourceGetter gets resources.
type ResourceGetter interface {
	GetCacheCR() runtime.Object
	GetResourceList() corev1.ResourceList
	GetServiceName() string
	GetService() *corev1.Service
	GetSecretName() string
	GetSecret() *corev1.Secret
	GetServerReplica() int
	GetClusterServerReplica() int
	GetStorageSize() string
}

type RedisResourceManager struct {
	cluster *v1alpha2.HarborCluster
}

const (
	defaultResourceCPU     = "1"
	defaultResourceMemory  = "2Gi"
	defaultResourceReplica = 3
	defaultStorageSize     = "1Gi"
)

const (
	labelApp = "goharbor.io/harbor-cluster"
)

// GetCacheCR gets cache cr instance.
func (rm *RedisResourceManager) GetCacheCR() runtime.Object {
	resource := rm.GetResourceList()
	pvc, _ := GenerateStoragePVC(rm.cluster.Name, rm.GetStorageSize(), rm.GetLabels())
	return &redisOp.RedisFailover{
		TypeMeta: metav1.TypeMeta{
			Kind:       redisOp.RFKind,
			APIVersion: "databases.spotahome.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      rm.cluster.Name,
			Namespace: rm.cluster.Namespace,
			Labels:    rm.GetLabels(),
		},
		Spec: redisOp.RedisFailoverSpec{
			Redis: redisOp.RedisSettings{
				Replicas: int32(rm.GetServerReplica()),
				Resources: corev1.ResourceRequirements{
					Limits:   resource,
					Requests: resource,
				},
				Storage: redisOp.RedisStorage{
					PersistentVolumeClaim: pvc,
				},
			},
			Sentinel: redisOp.SentinelSettings{
				Replicas: int32(rm.GetClusterServerReplica()),
				Resources: corev1.ResourceRequirements{
					Limits:   resource,
					Requests: resource,
				},
			},
			Auth: redisOp.AuthSettings{SecretPath: rm.GetSecretName()},
		},
	}
}

// GetServiceName gets service name.
func (rm *RedisResourceManager) GetServiceName() string {
	return fmt.Sprintf("%s-%s", "redis", rm.cluster.Name)
}

// GetService gets service.
func (rm *RedisResourceManager) GetService() *corev1.Service {
	name := rm.GetServiceName()
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: rm.cluster.Namespace,
			Labels:    rm.GetLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app.kubernetes.io/component": "redis",
				"app.kubernetes.io/name":      rm.cluster.Name,
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

// GetSecretName gets secret name.
func (rm *RedisResourceManager) GetSecretName() string {
	return fmt.Sprintf("%s-%s", "redis", rm.cluster.Name)
}

// GetSecret gets redis secret.
func (rm *RedisResourceManager) GetSecret() *corev1.Secret {
	name := rm.GetSecretName()
	passStr := common.RandomString(8, "a")
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: rm.cluster.Namespace,
			Labels:    rm.GetLabels(),
		},
		StringData: map[string]string{
			"redis-password": passStr,
			"password":       passStr,
		},
	}
}

// GetLabels gets labels merged from cluster labels.
func (rm *RedisResourceManager) GetLabels() map[string]string {
	dynLabels := map[string]string{
		"app.kubernetes.io/name":     "cache",
		"app.kubernetes.io/instance": rm.cluster.Namespace,
		labelApp:                     rm.cluster.Name,
	}
	return MergeLabels(dynLabels, rm.cluster.Labels)
}

// GetResourceList gets redis resources.
func (rm *RedisResourceManager) GetResourceList() corev1.ResourceList {
	resources := corev1.ResourceList{}
	if rm.cluster.Spec.InClusterCache.RedisSpec.Server == nil {
		resources, _ = GenerateResourceList(defaultResourceCPU, defaultResourceMemory)
		return resources
	}
	// assemble cpu
	if cpu := rm.cluster.Spec.InClusterCache.RedisSpec.Server.Resources.Requests.Cpu(); cpu != nil {
		resources[corev1.ResourceCPU] = *cpu
	}
	// assemble memory
	if mem := rm.cluster.Spec.InClusterCache.RedisSpec.Server.Resources.Requests.Memory(); mem != nil {
		resources[corev1.ResourceMemory] = *mem
	}
	return resources
}

// GetServerReplica gets deployment replica.
func (rm *RedisResourceManager) GetServerReplica() int {
	if rm.cluster.Spec.InClusterCache.RedisSpec.Server == nil || rm.cluster.Spec.InClusterCache.RedisSpec.Server.Replicas == 0 {
		return defaultResourceReplica
	}
	return rm.cluster.Spec.InClusterCache.RedisSpec.Server.Replicas
}

// GetClusterServerReplica gets deployment replica of sentinel mode.
func (rm *RedisResourceManager) GetClusterServerReplica() int {
	if rm.cluster.Spec.InClusterCache.RedisSpec.Sentinel == nil || rm.cluster.Spec.InClusterCache.RedisSpec.Sentinel.Replicas == 0 {
		return defaultResourceReplica
	}
	return rm.cluster.Spec.InClusterCache.RedisSpec.Sentinel.Replicas
}

// GetStorageSize gets storage size.
func (rm *RedisResourceManager) GetStorageSize() string {
	if rm.cluster.Spec.InClusterCache.RedisSpec.Server == nil || rm.cluster.Spec.InClusterCache.RedisSpec.Server.Storage == "" {
		return defaultStorageSize
	}
	return rm.cluster.Spec.InClusterCache.RedisSpec.Server.Storage
}

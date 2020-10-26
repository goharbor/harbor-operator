package cache

import (
	"fmt"
	"math/rand"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/lcm"
	redisCli "github.com/spotahome/redis-operator/api/redisfailover/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	labels1 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReidsType    = "rfr"
	SentinelType = "rfs"

	RoleName          = "harbor-cluster"
	RedisSentinelPort = "26379"
)

// GetRedisName returns the name for redis resources
func (redis *RedisReconciler) GetRedisName() string {
	return generateName(ReidsType, redis.GetHarborClusterName())
}

func generateName(typeName, metaName string) string {
	return fmt.Sprintf("%s-%s", typeName, metaName)
}

// GetRedisPassword is get redis password
func (redis *RedisReconciler) GetRedisPassword(secretName string) (string, error) {
	var redisPassWord string
	redisPassMap, err := redis.GetRedisSecret(secretName)
	if err != nil {
		return "", err
	}
	for k, v := range redisPassMap {
		if k == "password" {
			redisPassWord = string(v)
			return redisPassWord, nil
		}
	}
	return redisPassWord, nil
}

// GetRedisSecret returns the Redis Password Secret
func (redis *RedisReconciler) GetRedisSecret(secretName string) (map[string][]byte, error) {
	secret := &corev1.Secret{}

	err := redis.Client.Get(types.NamespacedName{Name: secretName, Namespace: redis.HarborCluster.Namespace}, secret)
	if err != nil {
		return nil, err
	}
	redisPw := secret.Data
	return redisPw, nil
}

// GetDeploymentPods returns the Redis Sentinel pod list
func (redis *RedisReconciler) GetDeploymentPods() (*appsv1.Deployment, *corev1.PodList, error) {
	deploy := &appsv1.Deployment{}
	name := fmt.Sprintf("%s-%s", "rfs", redis.HarborCluster.Name)

	err := redis.Client.Get(types.NamespacedName{Name: name, Namespace: redis.HarborCluster.Namespace}, deploy)
	if err != nil {
		return nil, nil, err
	}

	opts := &client.ListOptions{}
	set := labels1.SelectorFromSet(deploy.Spec.Selector.MatchLabels)
	opts.LabelSelector = set

	pod := &corev1.PodList{}
	err = redis.Client.List(opts, pod)
	if err != nil {
		redis.Log.Error(err, "fail to get pod.", "namespace", redis.HarborCluster.Namespace, "name", name)
		return nil, nil, err
	}
	return deploy, pod, nil
}

// GetStatefulSetPods returns the Redis Server pod list
func (redis *RedisReconciler) GetStatefulSetPods() (*appsv1.StatefulSet, *corev1.PodList, error) {
	sts := &appsv1.StatefulSet{}
	name := fmt.Sprintf("%s-%s", "rfr", redis.HarborCluster.Name)

	err := redis.Client.Get(types.NamespacedName{Name: name, Namespace: redis.HarborCluster.Namespace}, sts)
	if err != nil {
		return nil, nil, err
	}

	opts := &client.ListOptions{}
	set := labels1.SelectorFromSet(sts.Spec.Selector.MatchLabels)
	opts.LabelSelector = set

	pod := &corev1.PodList{}
	err = redis.Client.List(opts, pod)
	if err != nil {
		redis.Log.Error(err, "fail to get pod.", "namespace", redis.HarborCluster.Namespace, "name", name)
		return nil, nil, err
	}
	return sts, pod, nil
}

// GetServiceUrl returns the Redis Sentinel pod ip or service name
func (redis *RedisReconciler) GetSentinelServiceUrl(pods []corev1.Pod) string {
	var url string
	_, err := rest.InClusterConfig()
	if err != nil {
		randomPod := pods[rand.Intn(len(pods))]
		url = randomPod.Status.PodIP
	} else {
		url = fmt.Sprintf("%s-%s.%s.svc", "rfs", redis.GetHarborClusterName(), redis.HarborCluster.Namespace)
	}

	return url
}

// GetHarborClusterName returns harbor cluster name
func (redis *RedisReconciler) GetHarborClusterName() string {
	return redis.HarborCluster.Name
}

// GetHarborClusterNamespace returns harbor cluster namespace
func (redis *RedisReconciler) GetHarborClusterNamespace() string {
	return redis.HarborCluster.Namespace
}

// GetRedisResource returns redis resource
func (redis *RedisReconciler) GetRedisResource() corev1.ResourceList {
	resources := corev1.ResourceList{}

	if redis.HarborCluster.Spec.Cache.RedisSpec.Server == nil {
		return GenerateResourceList("1", "2Gi")
	}

	cpu := redis.HarborCluster.Spec.Cache.RedisSpec.Server.Resources.Requests.Cpu()
	mem := redis.HarborCluster.Spec.Cache.RedisSpec.Server.Resources.Requests.Memory()

	if cpu != nil {
		resources[corev1.ResourceCPU] = *cpu
	}
	if mem != nil {
		resources[corev1.ResourceMemory] = *mem
	}
	return resources
}

// GenerateResourceList returns resource list
func GenerateResourceList(cpu string, memory string) corev1.ResourceList {
	resources := corev1.ResourceList{}
	if cpu != "" {
		resources[corev1.ResourceCPU], _ = resource.ParseQuantity(cpu)
	}
	if memory != "" {
		resources[corev1.ResourceMemory], _ = resource.ParseQuantity(memory)
	}
	return resources
}

// GetRedisServerReplica returns redis server replicas
func (redis *RedisReconciler) GetRedisServerReplica() int32 {
	if redis.HarborCluster.Spec.Cache.RedisSpec.Server == nil {
		return 3
	}

	if redis.HarborCluster.Spec.Cache.RedisSpec.Server.Replicas == 0 {
		return 3
	}
	return int32(redis.HarborCluster.Spec.Cache.RedisSpec.Server.Replicas)
}

// GetRedisSentinelReplica returns redis sentinel replicas
func (redis *RedisReconciler) GetRedisSentinelReplica() int32 {

	if redis.HarborCluster.Spec.Cache.RedisSpec.Sentinel == nil {
		return 3
	}

	if redis.HarborCluster.Spec.Cache.RedisSpec.Sentinel.Replicas == 0 {
		return 3
	}
	return int32(redis.HarborCluster.Spec.Cache.RedisSpec.Sentinel.Replicas)
}

// GetRedisStorageSize returns redis server storage size
func (redis *RedisReconciler) GetRedisStorageSize() string {
	if redis.HarborCluster.Spec.Cache.RedisSpec.Server == nil {
		return "1Gi"
	}

	if redis.HarborCluster.Spec.Cache.RedisSpec.Server.Storage == "" {
		return "1Gi"
	}
	return redis.HarborCluster.Spec.Cache.RedisSpec.Server.Storage
}

// GetPodsStatus returns deleting  and current pod list
func (redis *RedisReconciler) GetPodsStatus(podArray []corev1.Pod) ([]corev1.Pod, []corev1.Pod) {
	deletingPods := make([]corev1.Pod, 0)
	currentPods := make([]corev1.Pod, 0, len(podArray))
	currentPodsByPhase := make(map[corev1.PodPhase][]corev1.Pod)

	for _, p := range podArray {
		if p.DeletionTimestamp != nil {
			deletingPods = append(deletingPods, p)
			continue
		}
		currentPods = append(currentPods, p)
		podsInPhase, ok := currentPodsByPhase[p.Status.Phase]
		if !ok {
			podsInPhase = []corev1.Pod{p}
		} else {
			podsInPhase = append(podsInPhase, p)
		}
		currentPodsByPhase[p.Status.Phase] = podsInPhase
	}
	return deletingPods, currentPods
}

// GenRedisConnURL returns harbor component redis secret
func (c *RedisConnect) GenRedisConnURL(component string) string {
	switch c.Schema {
	case RedisSentinelSchema:
		return c.genRedisSentinelConnURL(component)
	case RedisServerSchema:
		return c.genRedisServerConnURL(component)
	default:
		return ""
	}
}

// genRedisSentinelConnURL returns redis sentinel connection url
func (c *RedisConnect) genRedisSentinelConnURL(component string) string {

	hostInfo := GenHostInfo(c.Endpoints, c.Port)
	if c.Password != "" {
		return fmt.Sprintf("redis+sentinel://:%s@%s/mymaster/0", c.Password, hostInfo)
	}

	return fmt.Sprintf("redis+sentinel://%s/mymaster/0", hostInfo)
}

// genRedisServerConnURL returns redis server connection url
func (c *RedisConnect) genRedisServerConnURL(component string) string {
	hostInfo := GenHostInfo(c.Endpoints, c.Port)
	if component == HarborCore {
		return fmt.Sprintf("%s,100,%s", hostInfo[0], c.Password)
	}
	if c.Password != "" {
		return fmt.Sprintf("redis://:%s@%s/0", c.Password, hostInfo[0])
	}

	return fmt.Sprintf("redis://%s/0", hostInfo[0])
}

// GetRedisFailover returns RedisFailover object
func (redis *RedisReconciler) GetRedisFailover() (*redisCli.RedisFailover, error) {
	rf := &redisCli.RedisFailover{}
	err := redis.Client.Get(types.NamespacedName{Name: redis.HarborCluster.Name, Namespace: redis.HarborCluster.Namespace}, rf)
	if err != nil {
		return nil, err
	}

	return rf, nil
}

func cacheNotReadyStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(goharborv1.CacheReady).
		WithStatus(corev1.ConditionFalse).
		WithReason(reason).
		WithMessage(message)
}

func cacheUnknownStatus() *lcm.CRStatus {
	return lcm.New(goharborv1.CacheReady).
		WithStatus(corev1.ConditionUnknown)
}

func cacheReadyStatus(properties *lcm.Properties) *lcm.CRStatus {
	return lcm.New(goharborv1.CacheReady).
		WithStatus(corev1.ConditionTrue).
		WithReason("redis already ready").
		WithMessage("harbor component redis secrets are already create.").
		WithProperties(*properties)
}

// GetRedisServiceUrl returns the Redis server pod ip or service name
func (redis *RedisReconciler) GetRedisServiceUrl(pods []corev1.Pod) string {
	var url string
	randomPod := pods[rand.Intn(len(pods))]
	_, err := rest.InClusterConfig()
	if err != nil {
		url = randomPod.Status.PodIP
	} else {
		url = fmt.Sprintf("%s-%s.%s.svc", "cluster", redis.GetHarborClusterName(), redis.HarborCluster.Namespace)
	}

	return url
}

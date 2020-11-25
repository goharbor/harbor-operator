package cache

import (
	"errors"
	"fmt"
	"math/rand"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"

	rediscli "github.com/go-redis/redis"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	labels1 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Readiness reconcile will check Redis sentinel cluster if that has available.
// It does:
// - create redis connection pool
// - ping redis server
// - return redis properties if redis has available
func (rc *RedisController) Readiness(cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	var (
		client *rediscli.Client
		err    error
	)

	client, err = rc.GetInClusterRedisInfo(cluster)
	if err != nil {
		rc.Log.Error(err, "Fail to create redis client.",
			"namespace", cluster.Namespace, "name", cluster.Name)
		return cacheNotReadyStatus(ErrorGetRedisClient, err.Error()), err
	}

	defer client.Close()

	if err := client.Ping().Err(); err != nil {
		rc.Log.Error(err, "Fail to check Redis.",
			"namespace", cluster.Namespace, "name", cluster.Name)
		return cacheNotReadyStatus(ErrorCheckRedisHealth, err.Error()), err
	}

	rc.Log.Info("Redis already ready.",
		"namespace", cluster.Namespace, "name", cluster.Name)

	properties := lcm.Properties{}
	properties.Add(lcm.CachePropertyName, rc.generateRedisSpec())
	return cacheReadyStatus(&properties), nil
}

func (rc *RedisController) generateRedisSpec() *v1alpha2.ExternalRedisSpec {
	return &v1alpha2.ExternalRedisSpec{
		RedisHostSpec: harbormetav1.RedisHostSpec{
			Host: rc.ResourceManager.GetServiceName(),
			Port: 6379,
		},
		RedisCredentials: harbormetav1.RedisCredentials{
			PasswordRef: rc.ResourceManager.GetSecretName(),
		},
	}
}

// GetRedisPassword is get redis password
func (rc *RedisController) GetRedisPassword(secretName, namespace string) (string, error) {
	var redisPassWord string
	redisPassMap, err := rc.GetRedisSecret(secretName, namespace)
	if err != nil {
		return "", err
	}
	for k, v := range redisPassMap {
		if k == "redis-password" {
			redisPassWord = string(v)
			return redisPassWord, nil
		}
	}
	return redisPassWord, nil
}

// GetRedisSecret returns the Redis Password Secret
func (rc *RedisController) GetRedisSecret(secretName, namespace string) (map[string][]byte, error) {
	secret := &corev1.Secret{}

	err := rc.Client.Get(types.NamespacedName{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		return nil, err
	}
	redisPw := secret.Data
	return redisPw, nil
}

// GetDeploymentPods returns the Redis Sentinel pod list
func (rc *RedisController) GetDeploymentPods(name, namespace string) (*appsv1.Deployment, *corev1.PodList, error) {
	deploy := &appsv1.Deployment{}
	deployName := fmt.Sprintf("%s-%s", "rfs", name)

	err := rc.Client.Get(types.NamespacedName{Name: deployName, Namespace: namespace}, deploy)
	if err != nil {
		return nil, nil, err
	}

	opts := &client.ListOptions{}
	set := labels1.SelectorFromSet(deploy.Spec.Selector.MatchLabels)
	opts.LabelSelector = set

	pod := &corev1.PodList{}
	err = rc.Client.List(opts, pod)
	if err != nil {
		rc.Log.Error(err, "fail to get pod.", "namespace", namespace, "name", deployName)
		return nil, nil, err
	}
	return deploy, pod, nil
}

// GetStatefulSetPods returns the Redis Server pod list
func (rc *RedisController) GetStatefulSetPods(name, namespace string) (*appsv1.StatefulSet, *corev1.PodList, error) {
	sts := &appsv1.StatefulSet{}
	stsName := fmt.Sprintf("%s-%s", "rfr", name)

	err := rc.Client.Get(types.NamespacedName{Name: stsName, Namespace: namespace}, sts)
	if err != nil {
		return nil, nil, err
	}

	opts := &client.ListOptions{}
	set := labels1.SelectorFromSet(sts.Spec.Selector.MatchLabels)
	opts.LabelSelector = set

	pod := &corev1.PodList{}
	err = rc.Client.List(opts, pod)
	if err != nil {
		rc.Log.Error(err, "fail to get pod.", "namespace", namespace, "name", stsName)
		return nil, nil, err
	}
	return sts, pod, nil
}

// GetInClusterRedisInfo returns inCluster redis sentinel pool client
func (rc *RedisController) GetInClusterRedisInfo(cluster *v1alpha2.HarborCluster) (*rediscli.Client, error) {

	var client *rediscli.Client

	secret := rc.ResourceManager.GetSecret()
	password, err := rc.GetRedisPassword(secret.Name, secret.Namespace)
	if err != nil {
		return nil, err
	}

	_, sentinelPodList, err := rc.GetDeploymentPods(cluster.Name, cluster.Namespace)
	if err != nil {
		rc.Log.Error(err, "Fail to get deployment pods.")
		return nil, err
	}

	_, redisPodList, err := rc.GetStatefulSetPods(cluster.Name, cluster.Namespace)
	if err != nil {
		rc.Log.Error(err, "Fail to get deployment pods.")
		return nil, err
	}

	if len(sentinelPodList.Items) == 0 || len(redisPodList.Items) == 0 {
		rc.Log.Info("pod list is empty，pls wait.")
		return nil, errors.New("pod list is empty，pls wait")
	}

	spec := cluster.Spec.InClusterCache.RedisSpec
	switch spec.Schema {
	case SchemaRedisSentinel:
		sentinelPodArray := sentinelPodList.Items
		_, currentSentinelPods := rc.GetPodsStatus(sentinelPodArray)
		if len(currentSentinelPods) == 0 {
			return nil, errors.New("need to requeue")
		}
		endpoint := rc.GetSentinelServiceUrl(cluster.Name, cluster.Namespace, currentSentinelPods)
		connect := &RedisConnect{
			Endpoints: []string{endpoint},
			Port:      RedisSentinelConnPort,
			Password:  password,
			GroupName: RedisSentinelConnGroup,
		}
		rc.RedisConnect = connect
		client = connect.NewRedisPool()
	case SchemaRedisServer:
		redisPodArray := redisPodList.Items
		_, currentRedisPods := rc.GetPodsStatus(redisPodArray)
		if len(currentRedisPods) == 0 {
			return nil, errors.New("need to requeue")
		}
		endpoint := rc.GetRedisServiceUrl(cluster.Name, cluster.Namespace, currentRedisPods)
		connect := &RedisConnect{
			Endpoints: []string{endpoint},
			Port:      RedisRedisConnPort,
			Password:  password,
			GroupName: spec.GroupName,
			Schema:    SchemaRedisServer,
		}
		rc.RedisConnect = connect
		client = connect.NewRedisClient()
	}

	return client, nil
}

// GetPodsStatus returns deleting  and current pod list
func (rc *RedisController) GetPodsStatus(podArray []corev1.Pod) ([]corev1.Pod, []corev1.Pod) {
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

// GetServiceUrl returns the Redis Sentinel pod ip or service name
func (rc *RedisController) GetSentinelServiceUrl(name, namespace string, pods []corev1.Pod) string {
	var url string
	_, err := rest.InClusterConfig()
	if err != nil {
		randomPod := pods[rand.Intn(len(pods))]
		url = randomPod.Status.PodIP
	} else {
		url = fmt.Sprintf("%s-%s.%s.svc", "rfs", name, namespace)
	}

	return url
}

// GetRedisServiceUrl returns the Redis server pod ip or service name
func (rc *RedisController) GetRedisServiceUrl(name, namespace string, pods []corev1.Pod) string {
	var url string
	randomPod := pods[rand.Intn(len(pods))]
	_, err := rest.InClusterConfig()
	if err != nil {
		url = randomPod.Status.PodIP
	} else {
		url = fmt.Sprintf("%s-%s.%s.svc", "cluster", name, namespace)
	}

	return url
}

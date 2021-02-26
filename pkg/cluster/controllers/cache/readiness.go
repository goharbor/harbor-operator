package cache

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
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
// - return redis properties if redis has available.
func (rc *RedisController) Readiness(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	rc.Log.Info("Redis already ready.",
		"namespace", cluster.Namespace, "name", cluster.Name)

	properties := lcm.Properties{}
	properties.Add(lcm.CachePropertyName, rc.generateRedisSpec())

	return cacheReadyStatus(&properties), nil
}

func (rc *RedisController) generateRedisSpec() *v1alpha2.ExternalRedisSpec {
	return &v1alpha2.ExternalRedisSpec{
		RedisHostSpec: harbormetav1.RedisHostSpec{
			Host:              fmt.Sprintf("rfs-%s", rc.ResourceManager.GetCacheCRName()),
			Port:              RedisSentinelConnPort,
			SentinelMasterSet: RedisSentinelConnGroup,
		},
		RedisCredentials: harbormetav1.RedisCredentials{
			PasswordRef: rc.ResourceManager.GetSecretName(),
		},
	}
}

// GetRedisPassword is get redis password.
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

// GetRedisSecret returns the Redis Password Secret.
func (rc *RedisController) GetRedisSecret(secretName, namespace string) (map[string][]byte, error) {
	secret := &corev1.Secret{}

	err := rc.Client.Get(types.NamespacedName{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		return nil, err
	}

	redisPw := secret.Data

	return redisPw, nil
}

// GetDeploymentPods returns the Redis Sentinel pod list.
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

// GetStatefulSetPods returns the Redis Server pod list.
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

// GetPodsStatus returns deleting  and current pod list.
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

// GetSentinelServiceURL returns the Redis Sentinel pod ip or service name.
func (rc *RedisController) GetSentinelServiceURL(name, namespace string, pods []corev1.Pod) string {
	var url string

	_, err := rest.InClusterConfig()
	if err != nil {
		randomPod := pods[rand.Intn(len(pods))]
		url = randomPod.Status.PodIP
	} else {
		url = fmt.Sprintf("%s-%s.%s.svc.cluster.local", "rfs", name, namespace)
	}

	return url
}

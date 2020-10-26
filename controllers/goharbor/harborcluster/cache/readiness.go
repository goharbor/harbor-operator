package cache

import (
	"errors"
	"fmt"
	"strings"

	rediscli "github.com/go-redis/redis"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/lcm"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	HarborChartMuseum = "chartMuseum"
	HarborClair       = "clair"
	HarborJobService  = "jobService"
	HarborRegistry    = "registry"
	HarborCore        = "coreURL"
)

var (
	components = []string{
		HarborChartMuseum,
		HarborClair,
		HarborJobService,
		HarborRegistry,
		HarborCore,
	}
)

// Readiness reconcile will check Redis sentinel cluster if that has available.
// It does:
// - create redis connection pool
// - ping redis server
// - return redis properties if redis has available
func (redis *RedisReconciler) Readiness() (*lcm.CRStatus, error) {
	var (
		client *rediscli.Client
		err    error
	)

	switch redis.HarborCluster.Spec.Cache.Kind {
	case ExternalComponent:
		client, err = redis.GetExternalRedisInfo()
	case InClusterComponent:
		client, err = redis.GetInClusterRedisInfo()
	}

	if err != nil {
		redis.Log.Error(err, "Fail to create redis client.",
			"namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name)
		return cacheNotReadyStatus(GetRedisClientError, err.Error()), err
	}

	defer client.Close()

	if err := client.Ping().Err(); err != nil {
		redis.Log.Error(err, "Fail to check Redis.",
			"namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name)
		return cacheNotReadyStatus(CheckRedisHealthError, err.Error()), err
	}

	redis.Log.Info("Redis already ready.",
		"namespace", redis.HarborCluster.Namespace, "name", redis.HarborCluster.Name)

	properties := lcm.Properties{}
	for _, component := range components {
		url := redis.RedisConnect.GenRedisConnURL(component)
		secretName := fmt.Sprintf("%s-redis", strings.ToLower(component))
		propertyName := fmt.Sprintf("%sSecret", component)

		if err := redis.DeployComponentSecret(component, url, "", secretName); err != nil {
			return cacheNotReadyStatus(CreateComponentSecretError, err.Error()), err
		}

		properties.Add(propertyName, secretName)
	}

	return cacheReadyStatus(&properties), nil
}

// DeployComponentSecret deploy harbor component redis secret
func (redis *RedisReconciler) DeployComponentSecret(component, url, namespace, secretName string) error {
	secret := &corev1.Secret{}

	sc := redis.generateHarborCacheSecret(component, secretName, url, namespace)

	switch redis.HarborCluster.Spec.Cache.Kind {
	case ExternalComponent:
		if err := controllerutil.SetControllerReference(redis.HarborCluster, sc, redis.Scheme); err != nil {
			return err
		}
	case InClusterComponent:
		rf, err := redis.GetRedisFailover()
		if err != nil {
			return err
		}
		if err := controllerutil.SetControllerReference(rf, sc, redis.Scheme); err != nil {
			return err
		}
	}

	err := redis.Client.Get(types.NamespacedName{Name: secretName, Namespace: redis.HarborCluster.Namespace}, secret)
	if err != nil && kerr.IsNotFound(err) {
		redis.Log.Info("Creating Harbor Component Secret",
			"namespace", redis.HarborCluster.Namespace,
			"name", secretName,
			"component", component)
		return redis.Client.Create(sc)
	}

	return err
}

func (redis *RedisReconciler) GetExternalRedisInfo() (*rediscli.Client, error) {
	var (
		connect  *RedisConnect
		endpoint []string
		port     string
		client   *rediscli.Client
		err      error
		pw       string
	)
	spec := redis.HarborCluster.Spec.Cache.RedisSpec
	switch spec.Schema {
	case RedisSentinelSchema:
		if len(spec.Hosts) < 1 || spec.GroupName == "" {
			return nil, errors.New(".redis.spec.hosts or .redis.spec.groupName is invalid")
		}

		endpoint, port = GetExternalRedisHost(spec)

		if spec.SecretName != "" {
			pw, err = redis.GetExternalRedisPassword(spec)
		}

		connect = &RedisConnect{
			Endpoints: endpoint,
			Port:      port,
			Password:  pw,
			GroupName: spec.GroupName,
			Schema:    RedisSentinelSchema,
		}

		redis.RedisConnect = connect
		client = connect.NewRedisPool()
	case RedisServerSchema:
		if len(spec.Hosts) != 1 {
			return nil, errors.New(".redis.spec.hosts is invalid")
		}
		endpoint, port = GetExternalRedisHost(spec)

		if spec.SecretName != "" {
			pw, err = redis.GetExternalRedisPassword(spec)
		}

		connect = &RedisConnect{
			Endpoints: endpoint,
			Port:      port,
			Password:  pw,
			GroupName: spec.GroupName,
			Schema:    RedisServerSchema,
		}
		redis.RedisConnect = connect
		client = connect.NewRedisClient()
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetExternalRedisHost returns external redis host list and port
func GetExternalRedisHost(spec *goharborv1.RedisSpec) ([]string, string) {
	var (
		endpoint []string
		port     string
	)
	for _, host := range spec.Hosts {
		sp := host.Host
		endpoint = append(endpoint, sp)
		port = host.Port
	}
	return endpoint, port
}

// GetExternalRedisPassword returns external redis password
func (redis *RedisReconciler) GetExternalRedisPassword(spec *goharborv1.RedisSpec) (string, error) {

	pw, err := redis.GetRedisPassword(spec.SecretName)
	if err != nil {
		return "", err
	}

	return pw, err
}

// GetInClusterRedisInfo returns inCluster redis sentinel pool client
func (redis *RedisReconciler) GetInClusterRedisInfo() (*rediscli.Client, error) {

	var client *rediscli.Client

	password, err := redis.GetRedisPassword(redis.HarborCluster.Name)
	if err != nil {
		return nil, err
	}

	_, sentinelPodList, err := redis.GetDeploymentPods()
	if err != nil {
		redis.Log.Error(err, "Fail to get deployment pods.")
		return nil, err
	}

	_, redisPodList, err := redis.GetStatefulSetPods()
	if err != nil {
		redis.Log.Error(err, "Fail to get deployment pods.")
		return nil, err
	}

	if len(sentinelPodList.Items) == 0 || len(redisPodList.Items) == 0 {
		redis.Log.Info("pod list is empty，pls wait.")
		return nil, errors.New("pod list is empty，pls wait")
	}

	spec := redis.HarborCluster.Spec.Cache.RedisSpec
	switch spec.Schema {
	case RedisSentinelSchema:
		sentinelPodArray := sentinelPodList.Items
		_, currentSentinelPods := redis.GetPodsStatus(sentinelPodArray)
		if len(currentSentinelPods) == 0 {
			return nil, errors.New("need to requeue")
		}
		endpoint := redis.GetSentinelServiceUrl(currentSentinelPods)
		connect := &RedisConnect{
			Endpoints: []string{endpoint},
			Port:      RedisSentinelConnPort,
			Password:  password,
			GroupName: RedisSentinelConnGroup,
		}
		redis.RedisConnect = connect
		client = connect.NewRedisPool()
	case RedisServerSchema:
		redisPodArray := redisPodList.Items
		_, currentRedisPods := redis.GetPodsStatus(redisPodArray)
		if len(currentRedisPods) == 0 {
			return nil, errors.New("need to requeue")
		}
		endpoint := redis.GetRedisServiceUrl(currentRedisPods)
		connect := &RedisConnect{
			Endpoints: []string{endpoint},
			Port:      RedisRedisConnPort,
			Password:  password,
			GroupName: spec.GroupName,
			Schema:    RedisServerSchema,
		}
		redis.RedisConnect = connect
		client = connect.NewRedisClient()
	}

	return client, nil
}

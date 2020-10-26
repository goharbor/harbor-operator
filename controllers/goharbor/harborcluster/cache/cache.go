package cache

import (
	"context"

	"github.com/go-logr/logr"
	goharborv2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/k8s"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/lcm"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// RedisReconciler implement the Reconciler interface and lcm.Controller interface.
type RedisReconciler struct {
	HarborCluster *goharborv2.HarborCluster
	CXT           context.Context
	Client        k8s.Client
	Recorder      record.EventRecorder
	Log           logr.Logger
	DClient       k8s.DClient
	Scheme        *runtime.Scheme
	ExpectCR      *unstructured.Unstructured
	ActualCR      *unstructured.Unstructured
	Labels        map[string]string
	RedisConnect  *RedisConnect
}

// Reconciler implements the reconcile logic of redis service
func (redis *RedisReconciler) Reconcile() (*lcm.CRStatus, error) {
	redis.Labels = redis.NewLabels()
	redis.Client.WithContext(redis.CXT)
	redis.DClient.WithContext(redis.CXT)

	crdClient := redis.DClient.WithResource(redisFailoversGVR).WithNamespace(redis.HarborCluster.Namespace)

	if redis.HarborCluster.Spec.Cache.Kind == InClusterComponent {
		actualCR, err := crdClient.Get(redis.HarborCluster.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return redis.Provision()
		} else if err != nil {
			return cacheNotReadyStatus(GetRedisClientError, err.Error()), err
		}

		expectCR, err := redis.generateRedisCR()
		if err != nil {
			return cacheNotReadyStatus(GenerateRedisCrError, err.Error()), err
		}

		if err := controllerutil.SetControllerReference(redis.HarborCluster, expectCR, redis.Scheme); err != nil {
			return cacheNotReadyStatus(SetOwnerReferenceError, err.Error()), err
		}

		redis.ActualCR = actualCR
		redis.ExpectCR = expectCR

		crStatus, err := redis.Update(nil)
		if err != nil {
			return crStatus, err
		}
	}

	crStatus, err := redis.Readiness()
	if err != nil {
		return crStatus, err
	}

	return crStatus, nil
}

func (redis *RedisReconciler) Provision() (*lcm.CRStatus, error) {
	crStatus, err := redis.Deploy()
	if err != nil {
		return crStatus, err
	}
	return crStatus, nil
}

func (redis *RedisReconciler) Delete() (*lcm.CRStatus, error) {
	panic("implement me")
}

func (redis *RedisReconciler) Scale() (*lcm.CRStatus, error) {
	panic("implement me")
}

func (redis *RedisReconciler) Update(spec *goharborv2.HarborCluster) (*lcm.CRStatus, error) {
	crStatus, err := redis.RollingUpgrades()
	if err != nil {
		return crStatus, err
	}
	return crStatus, nil
}

package cache

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var redisFailoversGVR = redisOp.SchemeGroupVersion.WithResource(redisOp.RFNamePlural)

// NewRedisController is constructor for redis controller.
func NewRedisController(ctx context.Context, opts ...k8s.Option) lcm.Controller {
	ctrlOpts := &k8s.CtrlOptions{}

	for _, o := range opts {
		o(ctrlOpts)
	}

	return &RedisController{
		Ctx:             ctx,
		DClient:         ctrlOpts.DClient,
		Client:          ctrlOpts.Client,
		Log:             ctrlOpts.Log,
		Scheme:          ctrlOpts.Scheme,
		ResourceManager: NewResourceManager(),
	}
}

// RedisController implements lcm.Controller interface.
type RedisController struct {
	Ctx                context.Context
	DClient            k8s.DClient
	Client             k8s.Client
	Recorder           record.EventRecorder
	Log                logr.Logger
	Scheme             *runtime.Scheme
	RedisConnect       *RedisConnect
	ResourceManager    ResourceManager
	expectCR, actualCR runtime.Object
}

func (rc *RedisController) HealthChecker() lcm.HealthChecker {
	return &RedisHealthChecker{}
}

// Apply creates/updates/scales the resources, like kubernetes apply operation.
func (rc *RedisController) Apply(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	rc.DClient.WithContext(ctx)
	rc.Client.WithContext(ctx)

	rc.ResourceManager.WithCluster(cluster)

	crdClient := rc.DClient.WithResource(redisFailoversGVR).WithNamespace(cluster.Namespace)

	actualCR, err := crdClient.Get(rc.ResourceManager.GetCacheCRName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return rc.Deploy(cluster)
	} else if err != nil {
		return cacheNotReadyStatus(ErrorGetRedisClient, err.Error()), err
	}

	rc.actualCR = actualCR

	expectCR := rc.ResourceManager.GetCacheCR()
	if err := controllerutil.SetControllerReference(cluster, expectCR.(metav1.Object), rc.Scheme); err != nil {
		return cacheNotReadyStatus(ErrorSetOwnerReference, err.Error()), err
	}

	rc.expectCR = expectCR

	crStatus, err := rc.Update(cluster)
	if err != nil {
		return crStatus, err
	}

	return rc.Readiness(ctx, cluster)
}

// Delete...
func (rc *RedisController) Delete(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func (rc *RedisController) Upgrade(ctx context.Context, cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func cacheNotReadyStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(v1alpha2.CacheReady).
		WithStatus(corev1.ConditionFalse).
		WithReason(reason).
		WithMessage(message)
}

func cacheUnknownStatus() *lcm.CRStatus {
	return lcm.New(v1alpha2.CacheReady).
		WithStatus(corev1.ConditionUnknown)
}

func cacheReadyStatus(properties *lcm.Properties) *lcm.CRStatus {
	return lcm.New(v1alpha2.CacheReady).
		WithStatus(corev1.ConditionTrue).
		WithReason("redis already ready").
		WithMessage("harbor component redis secrets are already create.").
		WithProperties(*properties)
}

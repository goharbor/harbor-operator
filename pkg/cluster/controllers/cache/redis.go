package cache

import (
	"context"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var redisFailoversGVR = redisOp.SchemeGroupVersion.WithResource(redisOp.RFNamePlural)

// NewRedisController is constructor for redis controller.
func NewRedisController(opts ...k8s.Option) lcm.Controller {
	ctrlOpts := &k8s.CtrlOptions{}

	for _, o := range opts {
		o(ctrlOpts)
	}

	return &RedisController{
		DClient:         ctrlOpts.DClient,
		Client:          ctrlOpts.Client,
		Log:             ctrlOpts.Log,
		Scheme:          ctrlOpts.Scheme,
		ResourceManager: NewResourceManager(ctrlOpts.ConfigStore, ctrlOpts.Log, ctrlOpts.Scheme),
		ConfigStore:     ctrlOpts.ConfigStore,
	}
}

// RedisController implements lcm.Controller interface.
type RedisController struct {
	DClient         *k8s.DynamicClientWrapper
	Client          client.Client
	Recorder        record.EventRecorder
	Log             logr.Logger
	Scheme          *runtime.Scheme
	RedisConnect    *RedisConnect
	ResourceManager ResourceManager
	ConfigStore     *configstore.Store
}

// Apply creates/updates/scales the resources, like kubernetes apply operation.
func (rc *RedisController) Apply(ctx context.Context, cluster *goharborv1.HarborCluster, _ ...lcm.Option) (*lcm.CRStatus, error) {
	rc.ResourceManager.WithCluster(cluster)
	crdClient := rc.DClient.DynamicClient(ctx, k8s.WithResource(redisFailoversGVR), k8s.WithNamespace(cluster.Namespace))

	actualCR, err := crdClient.Get(rc.ResourceManager.GetCacheCRName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return rc.Deploy(ctx, cluster)
	} else if err != nil {
		return cacheNotReadyStatus(ErrorGetRedisClient, err.Error()), err
	}

	expectCR, err := rc.ResourceManager.GetCacheCR(ctx, cluster)
	if err != nil {
		return cacheNotReadyStatus(ErrorGenerateRedisCr, err.Error()), err
	}

	if err := controllerutil.SetControllerReference(cluster, expectCR.(metav1.Object), rc.Scheme); err != nil {
		return cacheNotReadyStatus(ErrorSetOwnerReference, err.Error()), err
	}

	crStatus, err := rc.Update(ctx, cluster, actualCR, expectCR)
	if err != nil {
		return crStatus, err
	}

	return rc.Readiness(ctx, cluster)
}

// Delete...
func (rc *RedisController) Delete(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	return nil, errors.Errorf("not implemented")
}

func (rc *RedisController) Upgrade(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	return nil, errors.Errorf("not implemented")
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

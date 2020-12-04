package harbor

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/kustomize/kstatus/status"
)

type Controller struct {
	KubeClient          k8s.Client
	Ctx                 context.Context
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	ComponentToCRStatus *sync.Map
}

func (harbor *Controller) Apply(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	harborCR := &v1alpha2.Harbor{}

	// Use the ctx from the parameter
	harbor.KubeClient.WithContext(ctx)

	nsdName := harbor.getHarborCRNamespacedName(harborcluster)

	err := harbor.KubeClient.Get(nsdName, harborCR)
	if err != nil {
		if errors.IsNotFound(err) {
			harborCR = harbor.getHarborCR(harborcluster)

			harbor.Log.Info("create harbor service", "name", nsdName)

			err = harbor.KubeClient.Create(harborCR)
			if err != nil {
				return harborClusterCRNotReadyStatus(CreateHarborCRError, err.Error()), err
			}
		} else {
			return harborClusterCRNotReadyStatus(GetHarborCRError, err.Error()), err
		}
	} else {
		harborCR = harbor.getHarborCR(harborcluster)

		// TODO: maybe we still need to do actual and desired status here to determine if we need to do update
		harbor.Log.Info("update harbor service", "name", nsdName)

		err = harbor.KubeClient.Update(harborCR)
		if err != nil {
			return harborClusterCRNotReadyStatus(UpdateHarborCRError, err.Error()), err
		}
	}

	harbor.Log.Info("harbor service is ready", "name", nsdName)

	return harborClusterCRStatus(harborCR), nil
}

func (harbor *Controller) Delete(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (harbor *Controller) Upgrade(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

// WithDependency appends the related dependent service for deploying Harbor later.
func (harbor *Controller) WithDependency(component v1alpha2.Component, svcCR *lcm.CRStatus) {
	harbor.ComponentToCRStatus.Store(component, svcCR)
}

func NewHarborController(ctx context.Context, options ...k8s.Option) *Controller {
	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}

	return &Controller{
		Ctx:                 ctx,
		KubeClient:          o.Client,
		Log:                 o.Log,
		Scheme:              o.Scheme,
		ComponentToCRStatus: &sync.Map{},
	}
}

// getHarborCR will get a Harbor CR from the harborcluster definition.
func (harbor *Controller) getHarborCR(harborcluster *v1alpha2.HarborCluster) *v1alpha2.Harbor {
	namespacedName := harbor.getHarborCRNamespacedName(harborcluster)

	harborCR := &v1alpha2.Harbor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			Labels: map[string]string{
				k8s.HarborClusterNameLabel: harborcluster.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(harborcluster, v1alpha2.HarborClusterGVK),
			},
		},
		Spec: harborcluster.Spec.HarborSpec,
	}

	// Use incluster spec in first priority.
	// Check based on the case that if the related dependent services are created
	if db := harbor.getDatabaseSpec(); db != nil {
		harbor.Log.Info("use incluster database", "database", db.Hosts)
		harborCR.Spec.Database = db
	}

	if cache := harbor.getCacheSpec(); cache != nil {
		harbor.Log.Info("use incluster cache", "cache", cache.Host)
		harborCR.Spec.Redis = cache
	}

	if storage := harbor.getStorageSpec(); storage != nil {
		harbor.Log.Info("use incluster storage", "storage", storage.S3.RegionEndpoint)
		harborCR.Spec.ImageChartStorage = storage
	}

	return harborCR
}

func harborClusterCRNotReadyStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(v1alpha2.ServiceReady).WithStatus(corev1.ConditionFalse).WithReason(reason).WithMessage(message)
}

func harborClusterCRUnknownStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(v1alpha2.ServiceReady).WithStatus(corev1.ConditionUnknown).WithReason(reason).WithMessage(message)
}

// harborClusterCRStatus will assembly the harbor cluster status according the v1alpha1.Harbor status.
func harborClusterCRStatus(harbor *v1alpha2.Harbor) *lcm.CRStatus {
	for _, condition := range harbor.Status.Conditions {
		if condition.Type == status.ConditionInProgress {
			return lcm.New(v1alpha2.ServiceReady).WithStatus(condition.Status).WithMessage(condition.Message).WithReason(condition.Reason)
		}
	}

	return harborClusterCRUnknownStatus(EmptyHarborCRStatusError, "The ready condition of harbor.goharbor.io is empty. Please wait for minutes.")
}

func (harbor *Controller) getHarborCRNamespacedName(harborcluster *v1alpha2.HarborCluster) types.NamespacedName {
	return types.NamespacedName{
		Namespace: harborcluster.Namespace,
		Name:      fmt.Sprintf("%s-harbor", harborcluster.Name),
	}
}

// getCacheSpec will get a name of k8s secret which stores cache info.
func (harbor *Controller) getCacheSpec() *v1alpha2.ExternalRedisSpec {
	p := harbor.getProperty(v1alpha2.ComponentCache, lcm.CachePropertyName)
	if p != nil {
		return p.Value.(*v1alpha2.ExternalRedisSpec)
	}

	return nil
}

// getDatabaseSecret will get a name of k8s secret which stores database info.
func (harbor *Controller) getDatabaseSpec() *v1alpha2.HarborDatabaseSpec {
	p := harbor.getProperty(v1alpha2.ComponentDatabase, lcm.DatabasePropertyName)
	if p != nil {
		return p.Value.(*v1alpha2.HarborDatabaseSpec)
	}

	return nil
}

// getStorageSecretForChartMuseum will get the secret name of chart museum storage config.
func (harbor *Controller) getStorageSpec() *v1alpha2.HarborStorageImageChartStorageSpec {
	p := harbor.getProperty(v1alpha2.ComponentStorage, lcm.StoragePropertyName)
	if p != nil {
		return p.Value.(*v1alpha2.HarborStorageImageChartStorageSpec)
	}

	return nil
}

func (harbor *Controller) getProperty(component v1alpha2.Component, name string) *lcm.Property {
	value, ok := harbor.ComponentToCRStatus.Load(component)
	if !ok {
		return nil
	}

	if value != nil {
		if crStatus, y := value.(*lcm.CRStatus); y {
			if len(crStatus.Properties) != 0 {
				return crStatus.Properties.Get(name)
			}
		}
	}

	return nil
}

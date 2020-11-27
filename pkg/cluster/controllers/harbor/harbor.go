package harbor

import (
	"context"
	"fmt"

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

type HarborController struct {
	KubeClient          k8s.Client
	Ctx                 context.Context
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	ComponentToCRStatus map[v1alpha2.Component]*lcm.CRStatus
}

func (harbor *HarborController) Apply(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	var harborCR *v1alpha2.Harbor
	err := harbor.KubeClient.Get(harbor.getHarborCRNamespacedName(harborcluster), harborCR)
	if err != nil {
		if errors.IsNotFound(err) {
			harborCR = harbor.getHarborCR(harborcluster)
			err = harbor.KubeClient.Create(harborCR)
			if err != nil {
				return harborClusterCRNotReadyStatus(CreateHarborCRError, err.Error()), err
			}
		} else {
			return harborClusterCRNotReadyStatus(GetHarborCRError, err.Error()), err
		}
	} else {
		harborCR = harbor.getHarborCR(harborcluster)
		err = harbor.KubeClient.Update(harborCR)
		if err != nil {
			return harborClusterCRNotReadyStatus(UpdateHarborCRError, err.Error()), err
		}
	}

	return harborClusterCRStatus(harborCR), nil
}

func (harbor *HarborController) Delete(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (harbor *HarborController) Upgrade(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func NewHarborController(ctx context.Context, options ...k8s.Option) *HarborController {
	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}
	return &HarborController{
		Ctx:        ctx,
		KubeClient: o.Client,
		Log:        o.Log,
		Scheme:     o.Scheme,
	}
}

// getHarborCR will get a Harbor CR from the harborcluster definition
func (harbor *HarborController) getHarborCR(harborcluster *v1alpha2.HarborCluster) *v1alpha2.Harbor {
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

	// use incluster spec in first priority
	if harborcluster.Spec.InClusterDatabase != nil {
		harborcluster.Spec.Database = harbor.getDatabaseSpec()
	}

	if harborcluster.Spec.InClusterCache != nil {
		harborcluster.Spec.Redis = harbor.getCacheSpec()
	}

	if harborcluster.Spec.InClusterStorage != nil {
		harborcluster.Spec.ImageChartStorage = *harbor.getStorageSpec()
	}

	return harborCR
}

func harborClusterCRNotReadyStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(v1alpha2.ServiceReady).WithStatus(corev1.ConditionFalse).WithReason(reason).WithMessage(message)
}

func harborClusterCRUnknownStatus(reason, message string) *lcm.CRStatus {
	return lcm.New(v1alpha2.ServiceReady).WithStatus(corev1.ConditionUnknown).WithReason(reason).WithMessage(message)
}

// harborClusterCRStatus will assembly the harbor cluster status according the v1alpha1.Harbor status
func harborClusterCRStatus(harbor *v1alpha2.Harbor) *lcm.CRStatus {
	for _, condition := range harbor.Status.Conditions {
		if condition.Type == status.ConditionInProgress {
			return lcm.New(v1alpha2.ServiceReady).WithStatus(condition.Status).WithMessage(condition.Message).WithReason(condition.Reason)
		}
	}
	return harborClusterCRUnknownStatus(EmptyHarborCRStatusError, "The ready condition of harbor.goharbor.io is empty. Please wait for minutes.")
}

func (harbor *HarborController) getHarborCRNamespacedName(harborcluster *v1alpha2.HarborCluster) types.NamespacedName {
	return types.NamespacedName{
		Namespace: harborcluster.Namespace,
		Name:      fmt.Sprintf("%s-harbor", harborcluster.Name),
	}
}

// getCacheSpec will get a name of k8s secret which stores cache info
func (harbor *HarborController) getCacheSpec() *v1alpha2.ExternalRedisSpec {
	p := harbor.getProperty(v1alpha2.ComponentCache, lcm.CachePropertyName)
	if p != nil {
		return p.Value.(*v1alpha2.ExternalRedisSpec)
	}
	return nil
}

// getDatabaseSecret will get a name of k8s secret which stores database info
func (harbor *HarborController) getDatabaseSpec() *v1alpha2.HarborDatabaseSpec {
	p := harbor.getProperty(v1alpha2.ComponentDatabase, lcm.DatabasePropertyName)
	if p != nil {
		return p.Value.(*v1alpha2.HarborDatabaseSpec)
	}
	return nil
}

// getStorageSecretForChartMuseum will get the secret name of chart museum storage config.
func (harbor *HarborController) getStorageSpec() *v1alpha2.HarborStorageImageChartStorageSpec {
	p := harbor.getProperty(v1alpha2.ComponentStorage, lcm.StoragePropertyName)
	if p != nil {
		return p.Value.(*v1alpha2.HarborStorageImageChartStorageSpec)
	}
	return nil
}

func (harbor *HarborController) getProperty(component v1alpha2.Component, name string) *lcm.Property {
	if harbor.ComponentToCRStatus == nil {
		return nil
	}
	crStatus := harbor.ComponentToCRStatus[component]
	if crStatus == nil || len(crStatus.Properties) == 0 {
		return nil
	}
	return crStatus.Properties.Get(name)
}

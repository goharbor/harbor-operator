package harbor

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"
)

type Controller struct {
	KubeClient client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
}

// Apply Harbor instance.
func (harbor *Controller) Apply(ctx context.Context, harborcluster *goharborv1.HarborCluster, options ...lcm.Option) (*lcm.CRStatus, error) {
	opts := &lcm.Options{}

	for _, op := range options {
		op(opts)
	}

	harborCR := &goharborv1.Harbor{}
	nsdName := harbor.getHarborCRNamespacedName(harborcluster)
	desiredCR := harbor.getHarborCR(ctx, harborcluster, opts.Dependencies)

	err := harbor.KubeClient.Get(ctx, nsdName, harborCR)
	if err != nil {
		if errors.IsNotFound(err) {
			harbor.Log.Info("Creating Harbor service")

			// Create a new one
			err = harbor.KubeClient.Create(ctx, desiredCR)
			if err != nil {
				return harborNotReadyStatus(CreateHarborCRError, err.Error()), err
			}

			harbor.Log.Info("Harbor service is created", "name", nsdName)

			return harborClusterCRStatus(harborCR), nil
		}

		// We don't know why none 404 error is returned, return unknown status
		return harborUnknownStatus(GetHarborCRError, err.Error()), err
	}

	// Found the existing one and check whether it needs to be updated
	if !common.Equals(ctx, harbor.Scheme, harborcluster, harborCR) {
		// Spec is changed, do update now
		harbor.Log.Info("Updating Harbor service", "name", nsdName)

		harborCR.Spec = desiredCR.Spec
		checksum.CopyMarkers(desiredCR, harborCR)

		if err := harbor.KubeClient.Update(ctx, harborCR); err != nil {
			return harborNotReadyStatus(UpdateHarborCRError, err.Error()), err
		}

		harbor.Log.Info("Harbor service is updated")
	}

	return harborClusterCRStatus(harborCR), nil
}

func (harbor *Controller) Delete(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (harbor *Controller) Upgrade(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func NewHarborController(options ...k8s.Option) *Controller {
	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}

	return &Controller{
		KubeClient: o.Client,
		Log:        o.Log,
		Scheme:     o.Scheme,
	}
}

// getHarborCR will get a Harbor CR from the harborcluster definition.
func (harbor *Controller) getHarborCR(ctx context.Context, harborcluster *goharborv1.HarborCluster, dependencies *lcm.CRStatusCollection) *goharborv1.Harbor { //nolint:funlen
	namespacedName := harbor.getHarborCRNamespacedName(harborcluster)

	spec := harborcluster.Spec.EmbeddedHarborSpec.DeepCopy()
	harborCR := &goharborv1.Harbor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			Labels: map[string]string{
				k8s.HarborClusterNameLabel: harborcluster.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(harborcluster, goharborv1.HarborClusterGVK),
			},
		},
		Spec: goharborv1.HarborSpec{
			ExternalURL: spec.ExternalURL,
			InternalTLS: goharborv1.HarborInternalTLSSpec{
				Enabled: spec.InternalTLS.Enabled,
			},
			ImageChartStorage:      &goharborv1.HarborStorageImageChartStorageSpec{},
			LogLevel:               spec.LogLevel,
			HarborAdminPasswordRef: spec.HarborAdminPasswordRef,
			UpdateStrategyType:     spec.UpdateStrategyType,
			Version:                spec.Version,
			Expose:                 spec.Expose,
			HarborComponentsSpec: goharborv1.HarborComponentsSpec{
				Portal:             spec.Portal,
				Core:               spec.Core,
				JobService:         spec.JobService,
				Registry:           spec.Registry,
				RegistryController: spec.RegistryController,
				ChartMuseum:        spec.ChartMuseum,
				Exporter:           spec.Exporter,
				Trivy:              spec.Trivy,
				Notary:             spec.Notary,
			},
			ImageSource: spec.ImageSource,
			Proxy:       spec.Proxy,
			Network:     harborcluster.Spec.Network,
			Trace:       harborcluster.Spec.Trace,
		},
	}

	if harborcluster.Spec.Storage.Spec.FileSystem != nil {
		harborCR.Spec.ImageChartStorage.FileSystem = harborcluster.Spec.Storage.Spec.FileSystem.HarborStorageImageChartStorageFileSystemSpec.DeepCopy()
	}

	if harborcluster.Spec.Storage.Spec.S3 != nil {
		harborCR.Spec.ImageChartStorage.S3 = harborcluster.Spec.Storage.Spec.S3.HarborStorageImageChartStorageS3Spec.DeepCopy()
	}

	if harborcluster.Spec.Storage.Spec.Swift != nil {
		harborCR.Spec.ImageChartStorage.Swift = harborcluster.Spec.Storage.Spec.Swift.HarborStorageImageChartStorageSwiftSpec.DeepCopy()
	}

	if harborcluster.Spec.Storage.Spec.Azure != nil {
		harborCR.Spec.ImageChartStorage.Azure = harborcluster.Spec.Storage.Spec.Azure.HarborStorageImageChartStorageAzureSpec.DeepCopy()
	}

	if harborcluster.Spec.Storage.Spec.Oss != nil {
		harborCR.Spec.ImageChartStorage.Oss = harborcluster.Spec.Storage.Spec.Oss.HarborStorageImageChartStorageOssSpec.DeepCopy()
	}

	if harborcluster.Spec.Storage.Spec.Gcs != nil {
		harborCR.Spec.ImageChartStorage.Gcs = harborcluster.Spec.Storage.Spec.Gcs.HarborStorageImageChartStorageGcsSpec.DeepCopy()
	}

	if harborcluster.Spec.Storage.Spec.Redirect != nil {
		harborCR.Spec.ImageChartStorage.Redirect.Disable = !harborcluster.Spec.Storage.Spec.Redirect.Enable
	}

	if harborcluster.Spec.Database.Spec.PostgreSQL != nil {
		harborCR.Spec.Database = harborcluster.Spec.Database.Spec.PostgreSQL.HarborDatabaseSpec.DeepCopy()
	}

	if harborcluster.Spec.Cache.Spec.Redis != nil {
		harborCR.Spec.Redis = harborcluster.Spec.Cache.Spec.Redis.DeepCopy()
	}

	// Use incluster spec in first priority.
	// Check based on the case that if the related dependent services are created
	if db := harbor.getDatabaseSpec(dependencies); db != nil {
		harbor.Log.Info("use incluster database", "database", db.Hosts)
		harborCR.Spec.Database = db
	}

	if cache := harbor.getCacheSpec(dependencies); cache != nil {
		harbor.Log.Info("use incluster cache", "cache", cache.Host, "port", cache.Port, "sentinelMasterSet", cache.SentinelMasterSet)
		harborCR.Spec.Redis = cache
	}

	if storage := harbor.getStorageSpec(dependencies); storage != nil {
		harbor.Log.Info("use incluster storage", "storage", storage.S3.RegionEndpoint)
		harborCR.Spec.ImageChartStorage = storage
	}

	// inject cert to harbor comps
	injectS3CertToHarborComponents(harborCR)

	dep := checksum.New(harbor.Scheme)
	dep.Add(ctx, harborcluster, true)
	dep.AddAnnotations(harborCR)

	return harborCR
}

func (harbor *Controller) getHarborCRNamespacedName(harborcluster *goharborv1.HarborCluster) types.NamespacedName {
	return types.NamespacedName{
		Namespace: harborcluster.Namespace,
		Name:      fmt.Sprintf("%s-harbor", harborcluster.Name),
	}
}

// getCacheSpec will get a name of k8s secret which stores cache info.
func (harbor *Controller) getCacheSpec(dependencies *lcm.CRStatusCollection) *goharborv1.ExternalRedisSpec {
	p := harbor.getProperty(dependencies, goharborv1.ComponentCache, lcm.CachePropertyName)
	if p != nil {
		return p.Value.(*goharborv1.ExternalRedisSpec)
	}

	return nil
}

// getDatabaseSecret will get a name of k8s secret which stores database info.
func (harbor *Controller) getDatabaseSpec(dependencies *lcm.CRStatusCollection) *goharborv1.HarborDatabaseSpec {
	p := harbor.getProperty(dependencies, goharborv1.ComponentDatabase, lcm.DatabasePropertyName)
	if p != nil {
		return p.Value.(*goharborv1.HarborDatabaseSpec)
	}

	return nil
}

// getStorageSecretForChartMuseum will get the secret name of chart museum storage config.
func (harbor *Controller) getStorageSpec(dependencies *lcm.CRStatusCollection) *goharborv1.HarborStorageImageChartStorageSpec {
	p := harbor.getProperty(dependencies, goharborv1.ComponentStorage, lcm.StoragePropertyName)
	if p != nil {
		return p.Value.(*goharborv1.HarborStorageImageChartStorageSpec)
	}

	return nil
}

func (harbor *Controller) getProperty(propertySet *lcm.CRStatusCollection, component goharborv1.Component, name string) *lcm.Property {
	crStatus, ok := propertySet.Get(component)
	if !ok {
		return nil
	}

	if len(crStatus.Properties) != 0 {
		return crStatus.Properties.Get(name)
	}

	return nil
}

func harborClusterCRStatus(harbor *goharborv1.Harbor) *lcm.CRStatus {
	var failedCondition, inProgressCondition *v1alpha1.Condition

	for _, condition := range harbor.Status.Conditions {
		if condition.Type == status.ConditionFailed {
			failedCondition = condition.DeepCopy()
		}

		if condition.Type == status.ConditionInProgress {
			inProgressCondition = condition.DeepCopy()
		}
	}

	if failedCondition == nil && inProgressCondition == nil {
		return harborUnknownStatus(EmptyHarborCRStatusError, "The ready condition of harbor.goharbor.io is empty. Please wait for minutes.")
	}

	if failedCondition != nil && failedCondition.Status == corev1.ConditionTrue {
		return harborNotReadyStatus(failedCondition.Reason, failedCondition.Message)
	}

	if inProgressCondition != nil && inProgressCondition.Status == corev1.ConditionTrue {
		return harborNotReadyStatus(inProgressCondition.Reason, inProgressCondition.Message)
	}

	return harborReadyStatus
}

// injectS3CertToHarborComponents injects s3 cert to harbor spec.
func injectS3CertToHarborComponents(harbor *goharborv1.Harbor) {
	storage := harbor.Spec.ImageChartStorage
	if storage == nil || storage.S3 == nil || storage.S3.CertificateRef == "" {
		return
	}

	certRef := storage.S3.CertificateRef
	// inject cert to component core
	harbor.Spec.Core.CertificateRefs = append(harbor.Spec.Core.CertificateRefs, certRef)
	// inject cert to component jobservice
	harbor.Spec.JobService.CertificateRefs = append(harbor.Spec.JobService.CertificateRefs, certRef)
	// inject cert to component trivy
	if harbor.Spec.Trivy != nil {
		harbor.Spec.Trivy.CertificateRefs = append(harbor.Spec.Trivy.CertificateRefs, certRef)
	}
	// inject cert to chartmuseum
	if harbor.Spec.ChartMuseum != nil {
		harbor.Spec.ChartMuseum.CertificateRefs = append(harbor.Spec.ChartMuseum.CertificateRefs, certRef)
	}
}

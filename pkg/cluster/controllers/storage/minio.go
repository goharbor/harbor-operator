package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	miniov2 "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/apis/minio.min.io/v2"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/ovh/configstore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Storage = "storage"

	DefaultCredsSecret  = "creds"
	DefaultPrefix       = "minio-"
	DefaultZone         = "zone-harbor"
	DefaultRegion       = "us-east-1"
	DefaultBucket       = "harbor"
	DefaultServicePort  = 80
	defaultMinIOService = "minio"
)

type MinIOController struct {
	KubeClient  client.Client
	Ctx         context.Context
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	MinioClient Minio
	ConfigStore *configstore.Store
}

var HarborClusterMinIOGVK = schema.GroupVersionKind{
	Group:   miniov2.SchemeGroupVersion.Group,
	Version: miniov2.SchemeGroupVersion.Version,
	Kind:    miniov2.MinIOCRDResourceKind,
}

func NewMinIOController(options ...k8s.Option) lcm.Controller {
	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}

	return &MinIOController{
		KubeClient:  o.Client,
		Log:         o.Log,
		Scheme:      o.Scheme,
		ConfigStore: o.ConfigStore,
	}
}

// Reconciler implements the reconcile logic of minIO service.
func (m *MinIOController) Apply(ctx context.Context, harborcluster *goharborv1.HarborCluster, _ ...lcm.Option) (*lcm.CRStatus, error) {
	// Apply minIO tenant
	if crs, err := m.applyTenant(ctx, harborcluster); err != nil {
		return crs, err
	}

	// Check readiness
	mt, tenantReady, err := m.checkMinIOReady(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(GetMinIOError, err.Error()), err
	}

	if !tenantReady {
		m.Log.Info("MinIO is not ready yet")

		return minioUnknownStatus(), nil
	}

	// Apply minIO ingress if necessary
	if crs, err := m.applyIngress(ctx, harborcluster); err != nil {
		return crs, err
	}

	// initializ bucket if necessary
	if crs, err := m.initializBucket(ctx, harborcluster, mt); crs != nil || err != nil {
		return crs, err
	}

	crs, err := m.ProvisionMinIOProperties(ctx, harborcluster, mt)
	if err != nil {
		return crs, err
	}

	m.Log.Info("MinIO is ready")

	return crs, nil
}

func (m *MinIOController) Delete(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	minioCR, err := m.generateMinIOCR(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(GenerateMinIOCrError, err.Error()), err
	}

	if err := m.KubeClient.Delete(ctx, minioCR); err != nil {
		return minioUnknownStatus(), err
	}

	return nil, nil
}

func (m *MinIOController) Upgrade(_ context.Context, _ *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	panic("implement me")
}

func (m *MinIOController) checkMinIOReady(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*miniov2.Tenant, bool, error) {
	minioCR := &miniov2.Tenant{}
	if err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), minioCR); err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	// For different version of minIO have different Status.
	// Ref https://github.com/minio/operator/commit/d387108ea494cf5cec57628c40d40604ac8d57ec#diff-48972613166d50a2acb9d562e33c5247
	if minioCR.Status.CurrentState == miniov2.StatusInitialized && minioCR.Status.AvailableReplicas == harborcluster.Spec.InClusterStorage.MinIOSpec.Replicas {
		ssName := fmt.Sprintf("%s-%s", m.getServiceName(harborcluster), DefaultZone)

		for _, pool := range minioCR.Status.Pools {
			if pool.SSName == ssName && pool.State == miniov2.PoolInitialized {
				return minioCR, true, nil
			}
		}
	}

	// Not ready
	return minioCR, false, nil
}

func (m *MinIOController) getMinIONamespacedName(harborcluster *goharborv1.HarborCluster) types.NamespacedName {
	return types.NamespacedName{
		Namespace: harborcluster.Namespace,
		Name:      m.getServiceName(harborcluster),
	}
}

func (m *MinIOController) getMinIOSecretNamespacedName(harborcluster *goharborv1.HarborCluster) types.NamespacedName {
	secretName := harborcluster.Spec.InClusterStorage.MinIOSpec.SecretRef
	if secretName == "" {
		secretName = DefaultPrefix + harborcluster.Name + "-" + DefaultCredsSecret
	}

	return types.NamespacedName{
		Namespace: harborcluster.Namespace,
		Name:      secretName,
	}
}

func (m *MinIOController) getServiceName(harborcluster *goharborv1.HarborCluster) string {
	return DefaultPrefix + harborcluster.Name
}

func (m *MinIOController) getTenantsServiceName(harborcluster *goharborv1.HarborCluster) string {
	// In latest minio operator, The name of the service is forced to be "minio"
	return defaultMinIOService
}

const (
	bucketInitializatedAnnotationKey = "minio.harbor.goharbor.io/bucket-initialized"
)

func (m *MinIOController) initializBucket(ctx context.Context, harborcluster *goharborv1.HarborCluster, tenant *miniov2.Tenant) (*lcm.CRStatus, error) {
	if m.isBucketInitialized(ctx, tenant) {
		return nil, nil
	}

	// Apply minio init job
	if crs, err := m.applyMinIOInitJob(ctx, harborcluster); err != nil {
		return crs, err
	}

	job, ready, err := m.checkMinIOInitJobReady(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(GetInitJobError, err.Error()), err
	}

	if !ready {
		m.Log.Info("MinIO init job is not ready yet")

		return minioUnknownStatus(), nil
	}

	if err := m.setBucketInitialized(ctx, tenant); err != nil {
		return minioNotReadyStatus(UpdateMinIOError, err.Error()), err
	}

	if err := m.KubeClient.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
		return minioNotReadyStatus(DeleteInitJobError, err.Error()), err
	}

	return nil, nil
}

func (m *MinIOController) isBucketInitialized(_ context.Context, tenant *miniov2.Tenant) bool {
	annotations := tenant.GetAnnotations()
	if annotations == nil {
		return false
	}

	s, ok := annotations[bucketInitializatedAnnotationKey]
	if !ok {
		return false
	}

	v, _ := strconv.ParseBool(s)

	return v
}

func (m *MinIOController) setBucketInitialized(ctx context.Context, tenant *miniov2.Tenant) error {
	annotations := tenant.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[bucketInitializatedAnnotationKey] = varTrueString

	tenant.SetAnnotations(annotations)

	return m.KubeClient.Update(ctx, tenant)
}

func minioNotReadyStatus(reason, message string) *lcm.CRStatus {
	now := metav1.Now()

	return &lcm.CRStatus{
		Condition: goharborv1.HarborClusterCondition{
			Type:               goharborv1.StorageReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: &now,
			Reason:             reason,
			Message:            message,
		},
		Properties: nil,
	}
}

func minioUnknownStatus() *lcm.CRStatus {
	now := metav1.Now()

	return &lcm.CRStatus{
		Condition: goharborv1.HarborClusterCondition{
			Type:               goharborv1.StorageReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: &now,
			Reason:             "",
			Message:            "",
		},
		Properties: nil,
	}
}

func minioReadyStatus(properties *lcm.Properties) *lcm.CRStatus {
	now := metav1.Now()

	return &lcm.CRStatus{
		Condition: goharborv1.HarborClusterCondition{
			Type:               goharborv1.StorageReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: &now,
			Reason:             "",
			Message:            "",
		},
		Properties: *properties,
	}
}

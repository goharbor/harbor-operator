package storage

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	miniov2 "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/apis/minio.min.io/v2"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/ovh/configstore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	// Set ownerReferences to MinIO PVS
	if err := m.setOwnerReferencesWithPVC(ctx, harborcluster); err != nil {
		return minioUnknownStatus(), err
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

	// Apply minio init job
	if crs, err := m.applyMinIOInitJob(ctx, harborcluster); err != nil {
		return crs, err
	}

	initJobReady, err := m.checkMinIOInitJobReady(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(GetInitJobError, err.Error()), err
	}

	if !initJobReady {
		m.Log.Info("MinIO init job is not ready yet")

		return minioUnknownStatus(), nil
	}

	crs, err := m.provisionMinIOProperties(ctx, harborcluster, mt)
	if err != nil {
		return crs, err
	}

	m.Log.Info("MinIO is ready")

	return crs, nil
}

func (m *MinIOController) setOwnerReferencesWithPVC(ctx context.Context, harborcluster *goharborv1.HarborCluster) (err error) {
	// Get the existing minIO CR first
	minioCR := &miniov2.Tenant{}
	if err = m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), minioCR); err != nil {
		return err
	}

	minioPVCs := &corev1.PersistentVolumeClaimList{}
	label := map[string]string{
		"v1.min.io/tenant": m.getServiceName(harborcluster),
	}

	opts := &client.ListOptions{}
	labelSelector := labels.SelectorFromSet(label)
	opts.LabelSelector = labelSelector

	if err = m.KubeClient.List(ctx, minioPVCs, opts); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	for i := range minioPVCs.Items {
		minioPVCs.Items[i].OwnerReferences = []metav1.OwnerReference{
			*metav1.NewControllerRef(minioCR, HarborClusterMinIOGVK),
		}

		err = m.KubeClient.Update(ctx, &minioPVCs.Items[i])
		if err != nil {
			m.Log.Error(err, "set pvc:", minioPVCs.Items[i].Name, "OwnerReferences error")
		}
	}

	return err
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
	if minioCR.Status.CurrentState == miniov2.StatusInitialized && minioCR.Status.AvailableReplicas == harborcluster.Spec.Storage.Spec.MinIO.Replicas {
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
	secretName := harborcluster.Spec.Storage.Spec.MinIO.SecretRef
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

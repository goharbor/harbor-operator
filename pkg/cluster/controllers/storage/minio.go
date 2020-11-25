package storage

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

const (
	Storage   = "storage"
	s3Storage = "s3"

	DefaultExternalSecretSuffix     = "harbor-cluster-storage"
	ChartMuseumExternalSecretSuffix = "chart-museum-storage"

	DefaultCredsSecret = "minio-creds"

	DefaultZone   = "zone-harbor"
	DefaultMinIO  = "minio"
	DefaultRegion = "us-east-1"
	DefaultBucket = "harbor"
)

type MinIOController struct {
	// TODO remove, use params harborcluster instead of HarborCluster.
	HarborCluster         *goharborv1.HarborCluster
	KubeClient            k8s.Client
	Ctx                   context.Context
	Log                   logr.Logger
	Scheme                *runtime.Scheme
	Recorder              record.EventRecorder
	CurrentMinIOCR        *minio.Tenant
	DesiredMinIOCR        *minio.Tenant
	CurrentExternalSecret *corev1.Secret
	DesiredExternalSecret *corev1.Secret
	MinioClient           Minio
}

var (
	HarborClusterMinIOGVK = schema.GroupVersionKind{
		Group:   minio.SchemeGroupVersion.Group,
		Version: minio.SchemeGroupVersion.Version,
		Kind:    minio.MinIOCRDResourceKind,
	}
)

func NewMinIOController(ctx context.Context, options ...k8s.Option) lcm.Controller {
	o := &k8s.CtrlOptions{}

	for _, option := range options {
		option(o)
	}
	return &MinIOController{
		Ctx:        ctx,
		KubeClient: o.Client,
		Log:        o.Log,
		Scheme:     o.Scheme,
	}
}

// Reconciler implements the reconcile logic of minIO service
func (m *MinIOController) Apply(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	var minioCR minio.Tenant

	m.HarborCluster = harborcluster
	m.DesiredMinIOCR = m.generateMinIOCR()

	err := m.KubeClient.Get(m.getMinIONamespacedName(), &minioCR)
	if k8serror.IsNotFound(err) {
		return m.Provision()
	} else if err != nil {
		return minioNotReadyStatus(GetMinIOError, err.Error()), err
	}

	m.CurrentMinIOCR = &minioCR

	// TODO remove scale event
	isScale, err := m.checkMinIOScale()
	if err != nil {
		return minioNotReadyStatus(ScaleMinIOError, err.Error()), err
	}
	if isScale {
		return m.Scale()
	}

	if m.checkMinIOUpdate() {
		return m.Update()
	}
	isReady, err := m.checkMinIOReady()
	if err != nil {
		return minioNotReadyStatus(GetMinIOError, err.Error()), err
	}

	if isReady {
		err := m.minioInit()
		if err != nil {
			return minioNotReadyStatus(CreateDefaultBucketError, err.Error()), err
		}
		return m.ProvisionMinIOProperties(&minioCR)
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) minioInit() error {
	accessKey, secretKey, err := m.getCredsFromSecret()
	if err != nil {
		return err
	}
	endpoint := m.getServiceName() + "." + m.HarborCluster.Namespace + ":9000"

	m.MinioClient, err = GetMinioClient(endpoint, string(accessKey), string(secretKey), DefaultRegion, false)
	if err != nil {
		return err
	}

	exists, err := m.MinioClient.IsBucketExists(DefaultBucket)
	if err != nil || exists {
		return err
	}

	err = m.MinioClient.CreateBucket(DefaultBucket)
	return err
}

func (m *MinIOController) checkMinIOUpdate() bool {
	return m.DesiredMinIOCR.Spec.Image != m.CurrentMinIOCR.Spec.Image
}

func (m *MinIOController) checkExternalUpdate() bool {
	return !cmp.Equal(m.DesiredExternalSecret.DeepCopy().Data, m.CurrentExternalSecret.DeepCopy().Data)
}

func (m *MinIOController) checkMinIOScale() (bool, error) {
	currentReplicas := m.CurrentMinIOCR.Spec.Zones[0].Servers
	desiredReplicas := m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Replicas
	if currentReplicas == desiredReplicas {
		return false, nil
	} else if currentReplicas == 1 {
		return false, fmt.Errorf("not support upgrading from standalone to distributed mode")
	}

	// MinIO creates erasure-coding sets of 4 to 16 drives per set.
	// The number of drives you provide in total must be a multiple of one of those numbers.
	// TODO validate by webhook
	if desiredReplicas%2 == 0 && desiredReplicas < 16 {
		return true, nil
	}

	return false, fmt.Errorf("for distributed mode, supply 4 to 16 drives (should be even)")
}

func (m *MinIOController) checkMinIOReady() (bool, error) {
	var minioCR minio.Tenant
	err := m.KubeClient.Get(m.getMinIONamespacedName(), &minioCR)

	// For different version of minIO have different Status.
	// Ref https://github.com/minio/operator/commit/d387108ea494cf5cec57628c40d40604ac8d57ec#diff-48972613166d50a2acb9d562e33c5247
	if minioCR.Status.CurrentState == minio.StatusReady || minioCR.Status.CurrentState == minio.StatusInitialized {
		return true, err
	}
	return false, err
}

func (m *MinIOController) getMinIONamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: m.HarborCluster.Namespace,
		Name:      m.getServiceName(),
	}
}

func (m *MinIOController) getMinIOSecretNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: m.HarborCluster.Namespace,
		Name:      m.HarborCluster.Name + "-" + DefaultCredsSecret,
	}
}

func (m *MinIOController) getExternalSecretNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: m.HarborCluster.Namespace,
		Name:      m.getExternalSecretName(),
	}
}

func (m *MinIOController) getExternalSecretName() string {
	return m.HarborCluster.Name + "-" + DefaultExternalSecretSuffix
}

func (m *MinIOController) getChartMuseumSecretName() string {
	return fmt.Sprintf("%s-%s", m.HarborCluster.Name, ChartMuseumExternalSecretSuffix)
}

func minioNotReadyStatus(reason, message string) *lcm.CRStatus {
	return &lcm.CRStatus{
		Condition: goharborv1.HarborClusterCondition{
			Type:               goharborv1.StorageReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             reason,
			Message:            message,
		},
		Properties: nil,
	}
}

func minioUnknownStatus() *lcm.CRStatus {
	return &lcm.CRStatus{
		Condition: goharborv1.HarborClusterCondition{
			Type:               goharborv1.StorageReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "",
			Message:            "",
		},
		Properties: nil,
	}
}

func minioReadyStatus(properties *lcm.Properties) *lcm.CRStatus {
	return &lcm.CRStatus{
		Condition: goharborv1.HarborClusterCondition{
			Type:               goharborv1.StorageReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "",
			Message:            "",
		},
		Properties: *properties,
	}
}

package storage

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	miniov2 "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/apis/minio.min.io/v2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	fsGroup            int64 = 10000
	runAsGroup         int64 = 10000
	runAsUser          int64 = 10000
	ComponentName            = "cluster-minio"
	ConfigImageKey           = "minio-docker-image"
	MinioInitName            = "cluster-minio-init"
	InitConfigImageKey       = "minio-init-docker-image"
)

func (m *MinIOController) ProvisionMinIOProperties(ctx context.Context, harborcluster *goharborv1.HarborCluster, minioInstance *miniov2.Tenant) (*lcm.CRStatus, error) {
	properties := &lcm.Properties{}

	data, err := m.getMinIOProperties(ctx, harborcluster, minioInstance)
	if err != nil {
		return minioNotReadyStatus(GetMinIOProperties, err.Error()), err
	}

	properties.Add(lcm.StoragePropertyName, data)

	return minioReadyStatus(properties), nil
}

func (m *MinIOController) getMinIOProperties(ctx context.Context, harborcluster *goharborv1.HarborCluster, minioInstance *miniov2.Tenant) (*goharborv1.HarborStorageImageChartStorageSpec, error) {
	accessKey, secretKey, err := m.getCredsFromSecret(ctx, harborcluster)
	if err != nil {
		return nil, err
	}

	secretKeyRef := m.createSecretKeyRef(secretKey, harborcluster, minioInstance)

	err = m.KubeClient.Create(ctx, secretKeyRef)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return nil, err
	}

	storageSpec := &goharborv1.HarborStorageImageChartStorageSpec{
		S3: &goharborv1.HarborStorageImageChartStorageS3Spec{
			RegistryStorageDriverS3Spec: goharborv1.RegistryStorageDriverS3Spec{
				AccessKey:    string(accessKey),
				SecretKeyRef: secretKeyRef.Name,
				Region:       DefaultRegion,
				Bucket:       DefaultBucket,
			},
		},
	}

	var (
		endpoint       string
		certificateRef string

		scheme     = corev1.URISchemeHTTP
		secure     = false
		v4Auth     = true
		skipVerify = true
	)

	if harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Enable {
		tls := harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.TLS
		if tls.Enabled() {
			secure = true
			skipVerify = false
			scheme = tls.GetScheme()
			certificateRef = tls.CertificateRef
		}

		endpoint = fmt.Sprintf("%s://%s", scheme, harborcluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.Ingress.Host)

		storageSpec.S3.CertificateRef = certificateRef
	} else {
		endpoint = fmt.Sprintf("http://%s.%s.svc:%d", m.getTenantsServiceName(harborcluster), harborcluster.Namespace, m.getServicePort())
	}

	storageSpec.S3.RegionEndpoint = strings.ToLower(endpoint)
	storageSpec.S3.Secure = &secure
	storageSpec.S3.V4Auth = &v4Auth
	storageSpec.S3.SkipVerify = skipVerify

	return storageSpec, nil
}

func (m *MinIOController) createSecretKeyRef(secretKey []byte, harborcluster *goharborv1.HarborCluster, minioInstance *miniov2.Tenant) *corev1.Secret {
	s3KeySecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "minio.min.io",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(harborcluster),
			Namespace:   harborcluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: m.generateAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(minioInstance, HarborClusterMinIOGVK),
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			v1alpha1.SharedSecretKey: secretKey,
		},
	}

	return s3KeySecret
}

// apply minIO tenant and its related service.
func (m *MinIOController) applyTenant(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	// If the expected tenant has been there
	minioCR := &miniov2.Tenant{}

	err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), minioCR)
	if k8serror.IsNotFound(err) {
		m.Log.Info("Creating minIO tenant")

		return m.createTenant(ctx, harborcluster)
	} else if err != nil {
		return minioNotReadyStatus(GetMinIOError, err.Error()), err
	}

	// Compare and do changes if necessary
	desiredMinIOCR, err := m.generateMinIOCR(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(GenerateMinIOCrError, err.Error()), err
	}

	if !common.Equals(ctx, m.Scheme, harborcluster, minioCR) {
		m.Log.Info("Updating minIO tenant")

		minioCR.Spec = *desiredMinIOCR.Spec.DeepCopy()
		checksum.CopyMarkers(desiredMinIOCR, minioCR)

		if err := m.KubeClient.Update(ctx, minioCR); err != nil {
			return minioNotReadyStatus(UpdateMinIOError, err.Error()), err
		}

		m.Log.Info("MinIO tenant is updated")
	}

	return minioUnknownStatus(), nil
}

// createTenant creates a new minio tenant.
func (m *MinIOController) createTenant(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	// If minio access secret is not specified, then create a random one.
	if harborcluster.Spec.InClusterStorage.MinIOSpec.SecretRef == "" {
		credsSecret := m.generateCredsSecret(harborcluster)

		err := m.KubeClient.Create(ctx, credsSecret)
		if err != nil && !k8serror.IsAlreadyExists(err) {
			return minioNotReadyStatus(CreateMinIOSecretError, err.Error()), err
		}
	}

	// Generate a desired CR
	desiredMinIOCR, err := m.generateMinIOCR(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(GenerateMinIOCrError, err.Error()), err
	}

	if err := m.KubeClient.Create(ctx, desiredMinIOCR); err != nil {
		return minioNotReadyStatus(CreateMinIOError, err.Error()), err
	}

	m.Log.Info("MinIO tenant is created")

	// Not confirm the final status yet, just return unknown status.
	return minioUnknownStatus(), nil
}

func (m *MinIOController) generateMinIOCR(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*miniov2.Tenant, error) { // nolint:funlen
	image, err := m.GetImage(ctx, harborcluster, ComponentName, ConfigImageKey)
	if err != nil {
		return nil, err
	}

	tenant := &miniov2.Tenant{
		TypeMeta: metav1.TypeMeta{
			Kind:       miniov2.MinIOCRDResourceKind,
			APIVersion: miniov2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(harborcluster),
			Namespace:   harborcluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: m.generateAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(harborcluster, goharborv1.HarborClusterGVK),
			},
		},
		Spec: miniov2.TenantSpec{
			// TODO soulseen: add sidecar container.
			Image:           image,
			ImagePullPolicy: m.getImagePullPolicy(ctx, harborcluster),
			ImagePullSecret: m.getImagePullSecret(ctx, harborcluster),
			Pools: []miniov2.Pool{
				{
					Name:                DefaultZone,
					Servers:             harborcluster.Spec.InClusterStorage.MinIOSpec.Replicas,
					VolumesPerServer:    harborcluster.Spec.InClusterStorage.MinIOSpec.VolumesPerServer,
					VolumeClaimTemplate: m.getVolumeClaimTemplate(harborcluster),
					Resources:           harborcluster.Spec.InClusterStorage.MinIOSpec.Resources,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup:    &fsGroup,
						RunAsGroup: &runAsGroup,
						RunAsUser:  &runAsUser,
					},
				},
			},
			Mountpath: miniov2.MinIOVolumeMountPath,
			CredsSecret: &corev1.LocalObjectReference{
				Name: m.getMinIOSecretNamespacedName(harborcluster).Name,
			},
			PodManagementPolicy: "Parallel",
			RequestAutoCert: func() *bool {
				b := false

				return &b
			}(),
			Env: []corev1.EnvVar{
				{
					Name:  "MINIO_BROWSER",
					Value: "on",
				},
			},
		},
	}

	dependencies := checksum.New(m.Scheme)
	dependencies.Add(ctx, harborcluster, true)
	dependencies.AddAnnotations(tenant)

	return tenant, nil
}

func (m *MinIOController) getVolumeClaimTemplate(harborcluster *goharborv1.HarborCluster) *corev1.PersistentVolumeClaim {
	isEmpty := reflect.DeepEqual(harborcluster.Spec.InClusterStorage.MinIOSpec.VolumeClaimTemplate, corev1.PersistentVolumeClaim{})
	if !isEmpty {
		return &harborcluster.Spec.InClusterStorage.MinIOSpec.VolumeClaimTemplate
	}

	defaultStorageClass := "default"

	return &corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &defaultStorageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}
}

func (m *MinIOController) getLabels() map[string]string {
	return map[string]string{"type": "harbor-cluster-minio", "app": "minio"}
}

func (m *MinIOController) generateAnnotations() map[string]string {
	// TODO
	return nil
}

func (m *MinIOController) generateCredsSecret(harborcluster *goharborv1.HarborCluster) *corev1.Secret {
	const SecretLen = 8
	credsAccesskey := common.RandomString(SecretLen, "a")
	credsSecretkey := common.RandomString(SecretLen, "a")

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "minio.min.io",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getMinIOSecretNamespacedName(harborcluster).Name,
			Namespace:   harborcluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: m.generateAnnotations(),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"accesskey": []byte(credsAccesskey),
			"secretkey": []byte(credsSecretkey),
		},
	}
}

func (m *MinIOController) getCredsFromSecret(ctx context.Context, harborcluster *goharborv1.HarborCluster) ([]byte, []byte, error) {
	var minIOSecret corev1.Secret

	namespaced := m.getMinIOSecretNamespacedName(harborcluster)
	err := m.KubeClient.Get(ctx, namespaced, &minIOSecret)

	return minIOSecret.Data["accesskey"], minIOSecret.Data["secretkey"], err
}

func (m *MinIOController) getImagePullPolicy(_ context.Context, harborcluster *goharborv1.HarborCluster) corev1.PullPolicy {
	if harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullPolicy != nil {
		return *harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullPolicy
	}

	if harborcluster.Spec.ImageSource != nil && harborcluster.Spec.ImageSource.ImagePullPolicy != nil {
		return *harborcluster.Spec.ImageSource.ImagePullPolicy
	}

	return config.DefaultImagePullPolicy
}

func (m *MinIOController) getImagePullSecret(_ context.Context, harborcluster *goharborv1.HarborCluster) corev1.LocalObjectReference {
	if len(harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullSecrets) > 0 {
		return harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullSecrets[0]
	}

	if harborcluster.Spec.ImageSource != nil && len(harborcluster.Spec.ImageSource.ImagePullSecrets) > 0 {
		return harborcluster.Spec.ImageSource.ImagePullSecrets[0]
	}

	return corev1.LocalObjectReference{Name: ""} // empty name means not using pull secret in minio
}

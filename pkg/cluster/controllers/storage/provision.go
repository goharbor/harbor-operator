package storage

import (
	"context"
	"fmt"
	"reflect"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	fsGroup    int64 = 10000
	runAsGroup int64 = 10000
	runAsUser  int64 = 10000
)

func (m *MinIOController) ProvisionMinIOProperties(minioInstamnce *minio.Tenant) (*lcm.CRStatus, error) {
	properties := &lcm.Properties{}

	data, err := m.getMinIOProperties(minioInstamnce)
	if err != nil {
		return minioNotReadyStatus(getMinIOProperties, err.Error()), err
	}

	properties.Add(lcm.StoragePropertyName, data)

	return minioReadyStatus(properties), nil
}

func (m *MinIOController) getMinIOProperties(minioInstance *minio.Tenant) (*goharborv1alpha2.HarborStorageImageChartStorageSpec, error) {
	accessKey, secretKey, err := m.getCredsFromSecret()
	if err != nil {
		return nil, err
	}

	secretKeyRef := m.createSecretKeyRef(secretKey, minioInstance)

	err = m.KubeClient.Create(secretKeyRef)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return nil, err
	}

	storageSpec := &goharborv1alpha2.HarborStorageImageChartStorageSpec{
		S3: &goharborv1alpha2.HarborStorageImageChartStorageS3Spec{
			RegistryStorageDriverS3Spec: goharborv1alpha2.RegistryStorageDriverS3Spec{
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

	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Enable {
		tls := m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.TLS
		if tls.Enabled() {
			secure = true
			skipVerify = false
			scheme = tls.GetScheme()
			certificateRef = tls.CertificateRef
		}

		endpoint = fmt.Sprintf("%s://%s", scheme, m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.Ingress.Host)

		storageSpec.S3.CertificateRef = certificateRef
	} else {
		endpoint = fmt.Sprintf("http://%s.%s.svc:%s", m.getServiceName(), m.HarborCluster.Namespace, "9000")
	}

	storageSpec.S3.RegionEndpoint = endpoint
	storageSpec.S3.Secure = &secure
	storageSpec.S3.V4Auth = &v4Auth
	storageSpec.S3.SkipVerify = skipVerify

	return storageSpec, nil
}

func (m *MinIOController) createSecretKeyRef(secretKey []byte, minioInstance *minio.Tenant) *corev1.Secret {
	s3KeySecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(),
			Namespace:   m.HarborCluster.Namespace,
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

func (m *MinIOController) Provision() (*lcm.CRStatus, error) {
	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.SecretRef == "" {
		credsSecret := m.generateCredsSecret()

		err := m.KubeClient.Create(credsSecret)
		if err != nil && !k8serror.IsAlreadyExists(err) {
			return minioNotReadyStatus(CreateMinIOSecretError, err.Error()), err
		}
	}

	err := m.KubeClient.Create(m.DesiredMinIOCR)
	if err != nil {
		return minioNotReadyStatus(CreateMinIOError, err.Error()), err
	}

	var minioCR minio.Tenant

	err = m.KubeClient.Get(m.getMinIONamespacedName(), &minioCR)
	if err != nil {
		return minioNotReadyStatus(CreateMinIOError, err.Error()), err
	}

	service := m.generateService()

	err = m.KubeClient.Create(service)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return minioNotReadyStatus(CreateMinIOServiceError, err.Error()), err
	}

	service.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(&minioCR, HarborClusterMinIOGVK),
	}

	err = m.KubeClient.Update(service)
	if err != nil {
		return minioNotReadyStatus(CreateMinIOServiceError, err.Error()), err
	}

	// expose minIO access endpoint by ingress.
	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Enable {
		ingress, err := m.generateIngress()
		if err != nil {
			return minioNotReadyStatus(CreateMinIOIngressError, err.Error()), err
		}

		err = m.KubeClient.Create(ingress)
		if err != nil && !k8serror.IsAlreadyExists(err) {
			return minioNotReadyStatus(CreateMinIOIngressError, err.Error()), err
		}

		ingress.OwnerReferences = []metav1.OwnerReference{
			*metav1.NewControllerRef(&minioCR, HarborClusterMinIOGVK),
		}

		err = m.KubeClient.Update(ingress)
		if err != nil {
			return minioNotReadyStatus(CreateMinIOServiceError, err.Error()), err
		}
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) generateIngress() (*netv1.Ingress, error) {
	var tls []netv1.IngressTLS

	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose != nil && m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.TLS.CertificateRef,
			Hosts:      []string{m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.Ingress.Host},
		}}
	}

	annotations := make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"

	if m.HarborCluster.Spec.Expose.Core.Ingress.Controller == v1alpha1.IngressControllerNCP {
		annotations["ncp/use-regex"] = "true"
		annotations["ncp/http-redirect"] = "true"
	}

	ingressPath, err := common.GetIngressPath(m.HarborCluster.Spec.Expose.Core.Ingress.Controller)

	return &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: netv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(),
			Namespace:   m.HarborCluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: annotations,
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{
				{
					Host: m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Expose.Ingress.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path: ingressPath,
									Backend: netv1.IngressBackend{
										ServiceName: m.getServiceName(),
										ServicePort: intstr.FromString("minio"),
									},
								},
							},
						},
					},
				},
			},
		},
	}, err
}

func (m *MinIOController) generateMinIOCR(ctx context.Context, harborcluster *goharborv1alpha2.HarborCluster) (*minio.Tenant, error) {
	image, err := m.GetImage(ctx, harborcluster)
	if err != nil {
		return nil, err
	}

	return &minio.Tenant{
		TypeMeta: metav1.TypeMeta{
			Kind:       minio.MinIOCRDResourceKind,
			APIVersion: minio.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(),
			Namespace:   m.HarborCluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: m.generateAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m.HarborCluster, goharborv1alpha2.HarborClusterGVK),
			},
		},
		Spec: minio.TenantSpec{
			SecurityContext: &corev1.PodSecurityContext{
				FSGroup:    &fsGroup,
				RunAsGroup: &runAsGroup,
				RunAsUser:  &runAsUser,
			},
			Metadata: &metav1.ObjectMeta{
				Labels:      m.getLabels(),
				Annotations: m.generateAnnotations(),
			},
			ServiceName:     m.getServiceName(),
			Image:           image,
			ImagePullPolicy: m.getImagePullPolicy(ctx, harborcluster),
			ImagePullSecret: m.getImagePullSecret(ctx, harborcluster),
			Zones: []minio.Zone{
				{
					Name:                DefaultZone,
					Servers:             m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Replicas,
					VolumesPerServer:    m.HarborCluster.Spec.InClusterStorage.MinIOSpec.VolumesPerServer,
					VolumeClaimTemplate: m.getVolumeClaimTemplate(),
					Resources:           *m.getResourceRequirements(),
				},
			},
			Mountpath: minio.MinIOVolumeMountPath,
			CredsSecret: &corev1.LocalObjectReference{
				Name: m.getMinIOSecretNamespacedName().Name,
			},
			PodManagementPolicy: "Parallel",
			RequestAutoCert:     false,
			CertConfig: &minio.CertificateConfig{
				CommonName:       "",
				OrganizationName: []string{},
				DNSNames:         []string{},
			},
			Env: []corev1.EnvVar{
				{
					Name:  "MINIO_BROWSER",
					Value: "on",
				},
			},
			Liveness: &minio.Liveness{
				InitialDelaySeconds: 120,
				PeriodSeconds:       60,
			},
		},
	}, nil
}

func (m *MinIOController) getServiceName() string {
	return DefaultPrefix + m.HarborCluster.Name
}

func (m *MinIOController) getServicePort() int32 {
	return 9000
}

func (m *MinIOController) getResourceRequirements() *corev1.ResourceRequirements {
	isEmpty := reflect.DeepEqual(m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Resources, corev1.ResourceRequirements{})
	if !isEmpty {
		return &m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Resources
	}

	limits := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("250m"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	requests := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("250m"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}

	return &corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func (m *MinIOController) getVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	isEmpty := reflect.DeepEqual(m.HarborCluster.Spec.InClusterStorage.MinIOSpec.VolumeClaimTemplate, corev1.PersistentVolumeClaim{})
	if !isEmpty {
		return &m.HarborCluster.Spec.InClusterStorage.MinIOSpec.VolumeClaimTemplate
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

func (m *MinIOController) generateService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getServiceName(),
			Namespace:   m.HarborCluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: m.generateAnnotations(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: m.getLabels(),
			Ports: []corev1.ServicePort{
				{
					Name:       "minio",
					Port:       m.getServicePort(),
					TargetPort: intstr.FromInt(9000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func (m *MinIOController) generateCredsSecret() *corev1.Secret {
	credsAccesskey := common.RandomString(8, "a")
	credsSecretkey := common.RandomString(8, "a")

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getMinIOSecretNamespacedName().Name,
			Namespace:   m.HarborCluster.Namespace,
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

func (m *MinIOController) getCredsFromSecret() ([]byte, []byte, error) {
	var minIOSecret corev1.Secret

	namespaced := m.getMinIOSecretNamespacedName()
	err := m.KubeClient.Get(namespaced, &minIOSecret)

	return minIOSecret.Data["accesskey"], minIOSecret.Data["secretkey"], err
}

func (m *MinIOController) getImagePullPolicy(_ context.Context, harborcluster *goharborv1alpha2.HarborCluster) corev1.PullPolicy {
	if harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullPolicy != nil {
		return *harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullPolicy
	}

	if harborcluster.Spec.ImageSource != nil && harborcluster.Spec.ImageSource.ImagePullPolicy != nil {
		return *harborcluster.Spec.ImageSource.ImagePullPolicy
	}

	return config.DefaultImagePullPolicy
}

func (m *MinIOController) getImagePullSecret(_ context.Context, harborcluster *goharborv1alpha2.HarborCluster) corev1.LocalObjectReference {
	if len(harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullSecrets) > 0 {
		return harborcluster.Spec.InClusterStorage.MinIOSpec.ImagePullSecrets[0]
	}

	if harborcluster.Spec.ImageSource != nil && len(harborcluster.Spec.ImageSource.ImagePullSecrets) > 0 {
		return harborcluster.Spec.ImageSource.ImagePullSecrets[0]
	}

	return corev1.LocalObjectReference{Name: ""} // empty name means not using pull secret in minio
}

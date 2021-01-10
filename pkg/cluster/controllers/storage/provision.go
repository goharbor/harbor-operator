package storage

import (
	"fmt"
	"reflect"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func (m *MinIOController) getMinIOProperties(minioInstance *minio.Tenant) (*goharborv1.HarborStorageImageChartStorageSpec, error) {
	accessKey, secretKey, err := m.getCredsFromSecret()
	if err != nil {
		return nil, err
	}

	secretKeyRef := m.createSecretKeyRef(secretKey, minioInstance)

	err = m.KubeClient.Create(secretKeyRef)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return nil, err
	}

	var (
		endpoint string
		scheme   corev1.URIScheme

		secure     = false
		v4Auth     = true
		skipVerify = true
	)

	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Disable {
		tls := m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.TLS
		if tls.Enabled() {
			secure = true
			skipVerify = false
			scheme = tls.GetScheme()
		}

		port := tls.GetInternalPort()

		endpoint = fmt.Sprintf("%s://%s:%d", scheme, m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Host, port)
	} else {
		endpoint = fmt.Sprintf("http://%s.%s.svc:%s", m.getServiceName(), m.HarborCluster.Namespace, "9000")
	}

	storageSpec := &goharborv1.HarborStorageImageChartStorageSpec{
		S3: &goharborv1.HarborStorageImageChartStorageS3Spec{
			RegistryStorageDriverS3Spec: goharborv1.RegistryStorageDriverS3Spec{
				AccessKey:      string(accessKey),
				SecretKeyRef:   secretKeyRef.Name,
				Region:         DefaultRegion,
				RegionEndpoint: endpoint,
				Bucket:         DefaultBucket,
				Secure:         &secure,
				V4Auth:         &v4Auth,
				SkipVerify:     skipVerify,
			},
		},
	}

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
	credsSecret := m.generateCredsSecret()

	err := m.KubeClient.Create(credsSecret)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return minioNotReadyStatus(CreateMinIOSecretError, err.Error()), err
	}

	err = m.KubeClient.Create(m.DesiredMinIOCR)
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
	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Disable {
		ingress := m.generateIngress()

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

func (m *MinIOController) generateIngress() *netv1.Ingress {
	var tls []netv1.IngressTLS

	if m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.TLS.CertificateRef,
			Hosts:      []string{m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Host},
		}}
	}

	annotations := make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"
	annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "HTTPS"

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
					Host: m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path: ingressPath,
									Backend: netv1.IngressBackend{
										ServiceName: m.getServiceName(),
										ServicePort: intstr.FromInt(9000),
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

func (m *MinIOController) generateMinIOCR() *minio.Tenant {
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
				*metav1.NewControllerRef(m.HarborCluster, goharborv1.HarborClusterGVK),
			},
		},
		Spec: minio.TenantSpec{
			Metadata: &metav1.ObjectMeta{
				Labels:      m.getLabels(),
				Annotations: m.generateAnnotations(),
			},
			ExternalCertSecret: &minio.LocalCertificateReference{
				Name: m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Redirect.TLS.CertificateRef,
				Type: "kubernetes.io/tls",
			},
			ServiceName: m.getServiceName(),
			Image:       m.GetImage(),
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
	}
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

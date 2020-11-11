package storage

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
	"strings"

	"github.com/goharbor/harbor-cluster-operator/controllers/common"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/storage/minio/api/v1"
	"github.com/goharbor/harbor-operator/pkg/lcm"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (m *MinIOController) ProvisionMinIOProperties(minioInstamnce *minio.Tenant) (*lcm.CRStatus, error) {
	properties := &lcm.Properties{}
	data,err := m.getMinIOProperties(minioInstamnce)
	if err != nil {
		return minioNotReadyStatus(getMinIOProperties, err.Error()), err
	}
	properties.Add(lcm.InClusterSecretForStorage, data)

	return minioReadyStatus(properties), nil
}

func (m *MinIOController) getMinIOProperties(minioInstance *minio.Tenant) (*goharborv1.RegistryStorageDriverS3Spec, error) {
	accessKey, secretKey, err := m.getCredsFromSecret()
	if err != nil {
		return nil, err
	}
	secretKeyRef := m.createSecretKeyRef(secretKey, minioInstance)
	err = m.KubeClient.Create(secretKeyRef)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return nil, err
	}
	var endpoint string
	if !m.HarborCluster.Spec.ImageChartStorage.Redirect.Disable {
		_, endpoint, err = GetMinIOHostAndSchema(m.HarborCluster.Spec.ExternalURL)
		if err != nil {
			return nil, err
		}
	} else {
		endpoint = fmt.Sprintf("http://%s.%s.svc:%s", m.getServiceName(), m.HarborCluster.Namespace, "9000")
	}

	secure := false
	v4Auth := false
	s3 := &goharborv1.RegistryStorageDriverS3Spec{
		AccessKey:      string(accessKey),
		SecretKeyRef:   secretKeyRef.Name,
		Region:         DefaultRegion,
		RegionEndpoint: endpoint,
		Bucket:         DefaultBucket,
		Secure:         &secure,
		V4Auth:         &v4Auth,
	}

	return s3, nil
}

func (m *MinIOController) createSecretKeyRef(secretKey []byte, minioInstance *minio.Tenant) *corev1.Secret {
	data := map[string]string{
		"secretkey": string(secretKey),
	}
	dataJson, _ := json.Marshal(&data)
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
			s3Storage: dataJson,
		},
	}
	return s3KeySecret
}

func (m *MinIOController) generateInClusterSecret(minioInstance *minio.Tenant) (inClusterSecret *corev1.Secret, chartMuseumSecret *corev1.Secret, err error) {
	labels := m.getLabels()
	labels[LabelOfStorageType] = inClusterStorage
	accessKey, secretKey, err := m.getCredsFromSecret()
	if err != nil {
		return nil, nil, err
	}

	var endpoint string
	if !m.HarborCluster.Spec.ImageChartStorage.Redirect.Disable {
		_, endpoint, err = GetMinIOHostAndSchema(m.HarborCluster.Spec.ExternalURL)
		if err != nil {
			return nil, nil, err
		}
	} else {
		endpoint = fmt.Sprintf("http://%s.%s.svc:%s", m.getServiceName(), m.HarborCluster.Namespace, "9000")
	}

	data := map[string]string{
		"accesskey":      string(accessKey),
		"secretkey":      string(secretKey),
		"region":         DefaultRegion,
		"bucket":         DefaultBucket,
		"regionendpoint": endpoint,
		"encrypt":        "false",
		"secure":         "false",
		"v4auth":         "false",
	}
	dataJson, _ := json.Marshal(&data)
	inClusterSecret = &corev1.Secret{
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
			s3Storage: dataJson,
		},
	}

	chartMuseumSecret = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.getChartMuseumSecretName(),
			Namespace:   m.HarborCluster.Namespace,
			Labels:      m.getLabels(),
			Annotations: m.generateAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(minioInstance, goharborv1.HarborClusterGVK),
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"kind":                  []byte("amazon"),
			"AWS_ACCESS_KEY_ID":     accessKey,
			"AWS_SECRET_ACCESS_KEY": secretKey,
			// use same bucket.
			"AMAZON_BUCKET":   []byte(DefaultBucket),
			"AMAZON_PREFIX":   []byte(fmt.Sprintf("%s-subfloder", DefaultBucket)),
			"AMAZON_REGION":   []byte(DefaultRegion),
			"AMAZON_ENDPOINT": []byte(endpoint),
		},
	}

	return inClusterSecret, chartMuseumSecret, nil
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

	// if disable redirect docker registry, we will expose minIO access endpoint by ingress.
	if !m.HarborCluster.Spec.ImageChartStorage.Redirect.Disable {
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

	service.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(&minioCR, HarborClusterMinIOGVK),
	}
	err = m.KubeClient.Update(service)
	if err != nil {
		return minioNotReadyStatus(CreateMinIOServiceError, err.Error()), err
	}

	return minioUnknownStatus(), nil
}

func GetMinIOHostAndSchema(accessURL string) (scheme string, host string, err error) {
	u, err := url.Parse(accessURL)
	if err != nil {
		return "", "", errors.Wrap(err, "invalid public URL")
	}

	hosts := strings.SplitN(u.Host, ":", 1)
	minioHost := "minio." + hosts[0]

	return u.Scheme, minioHost, nil
}

func (m *MinIOController) generateIngress() *netv1.Ingress {
	_, minioHost, err := GetMinIOHostAndSchema(m.HarborCluster.Spec.ExternalURL)
	if err != nil {
		panic(err)
	}

	var tls []netv1.IngressTLS

	if m.HarborCluster.Spec.Expose.Core.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: m.HarborCluster.Spec.Expose.Core.TLS.CertificateRef,
		}}
	}

	annotations := make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"

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
					Host: minioHost,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: netv1.IngressBackend{
										ServiceName: "minio",
										ServicePort: intstr.FromInt(9000),
									},
								},
							},
						},
					},
				},
			},
		},
	}
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
			ServiceName: m.getServiceName(),
			Image:       "minio/minio:" + m.HarborCluster.Spec.InClusterStorage.MinIOSpec.Version,
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
				corev1.EnvVar{
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
	return m.HarborCluster.Name + "-" + DefaultMinIO
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
			"secretkey": []byte(credsSecretkey)},
	}
}

func (m *MinIOController) getCredsFromSecret() ([]byte, []byte, error) {
	var minIOSecret corev1.Secret
	namespaced := m.getMinIOSecretNamespacedName()
	err := m.KubeClient.Get(namespaced, &minIOSecret)
	return minIOSecret.Data["accesskey"], minIOSecret.Data["secretkey"], err
}

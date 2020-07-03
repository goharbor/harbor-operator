package v1alpha2

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=harbor
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="h"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.externalURL`,description="The public URL to the Harbor application",priority=0
// Harbor is the Schema for the harbors API.
type Harbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HarborSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +resource:path=harbors
// HarborList contains a list of Harbor.
type HarborList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Harbor `json:"items"`
}

// HarborSpec defines the desired state of Harbor.
type HarborSpec struct {
	HarborHelm1_4_0Spec `json:",inline"`
}

type HarborHelm1_4_0Spec struct {
	HarborComponentsSpec `json:",inline"`

	// +kubebuilder:validation:Required
	Expose HarborExposeSpec `json:"expose"`

	// +kubebuilder:validation:Required
	ExternalURL string `json:"externalURL"`

	// +kubebuilder:validation:Optional
	InternalTLS HarborInternalTLSSpec `json:"internalTLS,omitempty"`

	// +kubebuilder:validation:Optional
	Persistence HarborPersistenceSpec `json:"persistence,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	LogLevel HarborLogLevel `json:"logLevel,omitempty"`

	// +kubebuilder:validation:Required
	HarborAdminPasswordRef string `json:"harborAdminPasswordRef"`

	// +kubebuilder:validation:Required
	// The secret key used for encryption.
	EncryptionKeyRef string `json:"encryptionKeyRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="RollingUpdate"
	UpdateStrategyType appsv1.DeploymentStrategyType `json:"updateStrategyType,omitempty"`

	// +kubebuilder:validation:Optional
	Proxy *CoreProxySpec `json:"proxy,omitempty"`
}

type HarborComponentsSpec struct {
	// +kubebuilder:validation:Required
	Portal ComponentSpec `json:"portal,omitempty"`

	// +kubebuilder:validation:Required
	Core CoreComponentSpec `json:"core,omitempty"`

	// +kubebuilder:validation:Required
	JobService JobServiceComponentSpec `json:"jobservice,omitempty"`

	// +kubebuilder:validation:Required
	Registry RegistryComponentSpec `json:"registry,omitempty"`

	// +kubebuilder:validation:Optional
	ChartMuseum *ChartMuseumComponentSpec `json:"chartmuseum,omitempty"`

	// +kubebuilder:validation:Optional
	Clair *ClairComponentSpec `json:"clair,omitempty"`

	// +kubebuilder:validation:Optional
	Trivy *TrivyComponentSpec `json:"trivy,omitempty"`

	// +kubebuilder:validation:Optional
	Notary *ComponentSpec `json:"notary,omitempty"`

	// +kubebuilder:validation:Optional
	// If null, redis dsn must be specified for every components that need a redis.
	Redis ExternalRedisSpec `json:"redis,omitempty"`
}

type ExternalRedisSpec struct {
	// +kubebuilder:validation:Required
	Address string `json:"address"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:default=6379
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	PasswordRef string `json:"passwordRef,omitempty"`
}

func (c *HarborComponentsSpec) RedisDSN(component ComponentWithRedis) OpacifiedDSN {
	switch component {
	case CoreRedis:
		if c.Core.Redis != nil {
			return *c.Core.Redis
		}
	case JobServiceRedis:
		if c.JobService.Redis != nil {
			return *c.JobService.Redis
		}
	case RegistryRedis:
		if c.Registry.Redis != nil {
			return *c.Registry.Redis
		}
	case ChartMuseumRedis:
		if c.ChartMuseum.Redis != nil {
			return *c.ChartMuseum.Redis
		}
	case ClairRedis:
		if c.Clair.Redis != nil {
			return *c.Clair.Redis
		}
	case TrivyRedis:
		if c.Trivy.Redis != nil {
			return *c.Trivy.Redis
		}
	}

	return OpacifiedDSN{
		DSN:         fmt.Sprintf("redis://%s:%d/%d", c.Redis.Address, c.Redis.Port, component.Index()),
		PasswordRef: c.Redis.PasswordRef,
	}
}

type CoreComponentSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// One of core redis dsn or global redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`

	// +kubebuilder:validation:Required
	// One of core database dsn or global database component must be specified
	Database CorePostgresqlSpec `json:"database,omitempty"`
}

type JobServiceComponentSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// One of jobservice redis dsn or global redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10
	WorkerCount int32 `json:"workerCount,omitempty"`
}

type RegistryComponentSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	RelativeURLs *bool `json:"relativeURLs,omitempty"`

	// +kubebuilder:validation:Optional
	StorageMiddlewares []RegistryMiddlewareSpec `json:"storageMiddlewares,omitempty"`

	// +kubebuilder:validation:Optional
	// One of redis dsn or redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`
}

type ChartMuseumComponentSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// One of chartmuseum redis dsn or global redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`
}

type ClairComponentSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// One of clair redis dsn or global redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`
}

type TrivyComponentSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// One of trivy redis dsn or global redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`
}

type HarborPersistenceSpec struct {
	// Setting it to "keep" to avoid removing PVCs during a helm delete
	// operation. Leaving it empty will delete PVCs after the chart deleted
	// +kubebuilder:default="keep"
	ResourcePolicy string `json:"resourcePolicy"`

	// +kubebuilder:validation:Optional
	PersistentVolumeClaim HarborPersistencePersistentVolumeClaimComponentsSpec `json:"persistentVolumeClaim,omitempty"`

	// +kubebuilder:validation:Optional
	ImageChartStorage HarborPersistenceImageChartStorageSpec `json:"imageChartStorage,omitempty"`
}

type HarborPersistenceImageChartStorageSpec struct {
	// +kubebuilder:validation:Optional
	Redirect RegistryStorageRedirectSpec `json:"redirect,omitempty"`

	// +kubebuilder:validation:Optional
	// FileSystem is an implementation of the storagedriver.StorageDriver interface which uses the local filesystem.
	// The local filesystem can be a remote volume.
	// See: https://docs.docker.com/registry/storage-drivers/filesystem/
	FileSystem *HarborPersistenceImageChartStorageFileSystemSpec `json:"filesystem,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Amazon S3 or S3 compatible services for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/s3/
	S3 *HarborPersistenceImageChartStorageS3Spec `json:"s3,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface that uses OpenStack Swift for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/swift/
	Swift *HarborPersistenceImageChartStorageSwiftSpec `json:"swift,omitempty"`
}

func (r *HarborPersistenceImageChartStorageSpec) Name() string {
	if r.S3 != nil {
		return "s3"
	}

	if r.Swift != nil {
		return "swift"
	}

	return "filesystem"
}

func (r *HarborPersistenceImageChartStorageSpec) ChartMuseum() ChartMuseumChartStorageDriverSpec {
	if r.S3 != nil {
		return ChartMuseumChartStorageDriverSpec{
			Amazon: r.S3.ChartMuseum(),
		}
	}

	if r.Swift != nil {
		return ChartMuseumChartStorageDriverSpec{
			OpenStack: r.Swift.ChartMuseum(),
		}
	}

	if r.FileSystem != nil {
		return ChartMuseumChartStorageDriverSpec{
			FileSystem: r.FileSystem.ChartMuseum(),
		}
	}

	return ChartMuseumChartStorageDriverSpec{
		FileSystem: &ChartMuseumChartStorageDriverFilesystemSpec{
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func (r *HarborPersistenceImageChartStorageSpec) Registry() RegistryStorageDriverSpec {
	if r.S3 != nil {
		return RegistryStorageDriverSpec{
			S3: &r.S3.RegistryStorageDriverS3Spec,
		}
	}

	if r.Swift != nil {
		return RegistryStorageDriverSpec{
			Swift: &r.Swift.RegistryStorageDriverSwiftSpec,
		}
	}

	if r.FileSystem != nil {
		return RegistryStorageDriverSpec{
			FileSystem: &RegistryStorageDriverFilesystemSpec{
				VolumeSource: corev1.VolumeSource{},
				MaxThreads:   r.FileSystem.MaxThreads,
			},
		}
	}

	return RegistryStorageDriverSpec{
		FileSystem: &RegistryStorageDriverFilesystemSpec{
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func (r *HarborPersistenceImageChartStorageSpec) Validate() error {
	found := 0

	if r.FileSystem != nil {
		found++
	}

	if r.S3 != nil {
		found++
	}

	if r.Swift != nil {
		found++
	}

	switch found {
	case 0:
		return ErrNoStorageConfiguration
	case 1:
		return nil
	default:
		return Err2StorageConfiguration
	}
}

type HarborPersistenceImageChartStorageFileSystemSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/storage"
	RootDirectory string `json:"rootDirectory,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=25
	// +kubebuilder:default=100
	MaxThreads int32 `json:"maxthreads,omitempty"`
}

func (r *HarborPersistenceImageChartStorageFileSystemSpec) ChartMuseum() *ChartMuseumChartStorageDriverFilesystemSpec {
	return &ChartMuseumChartStorageDriverFilesystemSpec{
		VolumeSource: corev1.VolumeSource{},
	}
}

type HarborPersistenceImageChartStorageS3Spec struct {
	RegistryStorageDriverS3Spec `json:",inline"`
}

func (r *HarborPersistenceImageChartStorageS3Spec) ChartMuseum() *ChartMuseumChartStorageDriverAmazonSpec {
	return &ChartMuseumChartStorageDriverAmazonSpec{
		AccessKeyID:     r.AccessKey,
		AccessSecretRef: r.SecretKeyRef,
		Bucket:          r.Bucket,
		Endpoint:        r.RegionEndpoint,
		Prefix:          r.RootDirectory,
		Region:          r.Region,
	}
}

type HarborPersistenceImageChartStorageSwiftSpec struct {
	RegistryStorageDriverSwiftSpec `json:",inline"`
}

func (r *HarborPersistenceImageChartStorageSwiftSpec) ChartMuseum() *ChartMuseumChartStorageDriverOpenStackSpec {
	return &ChartMuseumChartStorageDriverOpenStackSpec{
		AuthenticationURL: r.AuthenticationURL,
		Container:         r.Container,
		Domain:            r.Domain,
		DomainID:          r.DomainID,
		PasswordRef:       r.PasswordRef,
		Prefix:            r.Prefix,
		Region:            r.Region,
		Tenant:            r.Tenant,
		TenantID:          r.TenantID,
		Username:          r.Username,
	}
}

type HarborPersistencePersistentVolumeClaimComponentsSpec struct {
	// +kubebuilder:validation:Optional
	Registry HarborPersistencePersistentVolumeClaim5GSpec `json:"registry,omitempty"`

	// +kubebuilder:validation:Optional
	ChartMuseum HarborPersistencePersistentVolumeClaim5GSpec `json:"chartmuseum,omitempty"`

	// +kubebuilder:validation:Optional
	JobService HarborPersistencePersistentVolumeClaim1GSpec `json:"jobservice,omitempty"`

	// +kubebuilder:validation:Optional
	Database HarborPersistencePersistentVolumeClaim1GSpec `json:"database,omitempty"`

	// +kubebuilder:validation:Optional
	Redis HarborPersistencePersistentVolumeClaim1GSpec `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	Trivy HarborPersistencePersistentVolumeClaim5GSpec `json:"trivy,omitempty"`
}

type HarborPersistencePersistentVolumeClaim5GSpec struct {
	HarborPersistencePersistentVolumeClaimComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=5368709120
	Size int64 `json:"size,omitempty"`
}

type HarborPersistencePersistentVolumeClaim1GSpec struct {
	HarborPersistencePersistentVolumeClaimComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1073741824
	Size int64 `json:"size"`
}

type HarborPersistencePersistentVolumeClaimComponentSpec struct {
	// +kubebuilder:validation:Optional
	// Use the existing PVC which must be created manually before bound,
	// and specify the "subPath" if the PVC is shared with other components
	ExistingClaim string `json:"existingClaim,omitempty"`

	// +kubebuilder:validation:Optional
	// Specify the "storageClass" used to provision the volume.
	// Or the default StorageClass will be used(the default).
	// Set it to "-" to disable dynamic provisioning
	StorageClass string `json:"storageClass,omitempty"`

	// +kubebuilder:validation:Optional
	SubPath string `json:"subPath,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="ReadWriteOnce"
	AccessMode corev1.PersistentVolumeAccessMode `json:"accessMode,omitempty"`
}

type HarborInternalTLSSpec struct {
	// +kubebuilder:validation:Optional
	CoreCertificateRef string `json:"coreCertificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	JobServiceCertificateRef string `json:"jobServiceCertificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	RegistryCertificateRef string `json:"registryCertificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	PortalCertificateRef string `json:"portalCertificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	ChartMuseumCertificateRef string `json:"chartmuseumCertificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	ClairCertificateRef string `json:"clairCertificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	TrivyCertificateRef string `json:"trivyCertificateRef,omitempty"`
}

type HarborExposeSpec struct {
	// +kubebuilder:validation:Optional
	TLS HarborExposeTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	Ingress *HarborExposeIngressSpec `json:"ingress,omitempty"`

	// +kubebuilder:validation:Optional
	ClusterIP *HarborExposeClusterIPSpec `json:"clusterIP,omitempty"`

	// +kubebuilder:validation:Optional
	LoadBalancer *HarborExposeLoadBalancerSpec `json:"loadbalancer,omitempty"`
}

// Enables TLS for public traffic.
type HarborExposeTLSSpec struct {
	// +kubebuilder:validation:Required
	// CertificateRef is a reference to the secret containing public certificates.
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Optional
	// NotaryCertificateRef is a reference to the secret containing public Notary certificates.
	// Otherwise it will be the same values than certificateRef.
	NotaryCertificateRef string `json:"notaryCertificateRef,omitempty"`
}

type HarborExposeNodePortSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:default="harbor"
	// The name of NodePort service
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	Ports HarborExposeNodePortPortsSpec `json:"ports,omitempty"`
}

type HarborExposeLoadBalancerSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:default="harbor"
	// The name of NodePort service
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	IP string `json:"ip,omitempty"`

	// +kubebuilder:validation:Optional
	Ports HarborExposePortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +kubebuilder:validation:Optional
	SourceRanges []string `json:"sourceRanges,omitempty"`
}

type HarborExposeNodePortPortsSpec struct {
	// +kubebuilder:validation:Optional
	HTTP HarborExposeNodePortPortsHTTPSpec `json:"http,omitempty"`

	// +kubebuilder:validation:Optional
	HTTPS HarborExposeNodePortPortsHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Optional
	Notary HarborExposeNodePortPortsNotarySpec `json:"notary,omitempty"`
}

type HarborExposeNodePortPortsHTTPSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=80
	// The service port Harbor listens on when serving with HTTP
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=30002
	// The node port Harbor listens on when serving with HTTP
	NodePort int32 `json:"nodePort,omitempty"`
}

type HarborExposeNodePortPortsHTTPSSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=443
	// The service port Harbor listens on when serving with HTTPS
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=30003
	// The node port Harbor listens on when serving with HTTPS
	NodePort int32 `json:"nodePort,omitempty"`
}

type HarborExposeNodePortPortsNotarySpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=4443
	// The service port Notary listens on
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=30004
	// The node port Notary listens on
	NodePort int32 `json:"nodePort,omitempty"`
}

type HarborExposeClusterIPSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:default="harbor"
	// The name of ClusterIP service
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Optional
	Ports HarborExposePortsSpec `json:"ports,omitempty"`
}

type HarborExposePortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=80
	// The service port Harbor listens on when serving with HTTP.
	HTTPPort int32 `json:"httpPort,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=443
	// The service port Harbor listens on when serving with HTTPS.
	HTTPSPort int32 `json:"httpsPort,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMaximum=true
	// +kubebuilder:default=4443
	// The service port Notary listens on.
	// Only needed when notary is enabled.
	NotaryPort int32 `json:"notaryPort,omitempty"`
}

type HarborExposeIngressSpec struct {
	// +kubebuilder:validation:Required
	Hosts HarborExposeIngressHostsSpec `json:"hosts"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	Controller string `json:"controller,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"ingress.kubernetes.io/ssl-redirect":"true","ingress.kubernetes.io/proxy-body-size":"0","nginx.ingress.kubernetes.io/ssl-redirect":"true","nginx.ingress.kubernetes.io/proxy-body-size":"0"}
	Annotations map[string]string `json:"annotations,omitempty"`
}

type HarborExposeIngressHostsSpec struct {
	// +kubebuilder:validation:Required
	Core string `json:"core"`

	// +kubebuilder:validation:Required
	Notary string `json:"notary"`
}

type NotarySignerComponent struct {
	// CommonName is a common name to be used on the Certificate.
	// The CommonName should have a length of 64 characters or fewer to avoid
	// generating invalid CSRs.
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Optional
	CommonName string `json:"commonName,omitempty"`

	// Organization is the organization to be used on the Certificate
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// This cannot be set to true: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#validation
	// +listType:atomic
	Organization []string `json:"organization,omitempty"`

	// KeySize is the key bit size of the corresponding private key for this certificate.
	// +optional
	// +kubebuilder:validation:Maximum=8192
	// +kubebuilder:validation:Minimum=2048
	KeySize int32 `json:"keySize,omitempty"`
}

type ClairAdapterComponent struct {
	// +kubebuilder:validation:Required
	RedisSecret string `json:"redisSecret"`
}

type ClairComponent struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +listType:set
	VulnerabilitySources []string `json:"vulnerabilitySources"`

	// +kubebuilder:validation:Required
	Adapter ClairAdapterComponent `json:"adapter"`
}
type HarborComponents struct {
	Clair *ClairComponent `json:"clair"`

	NotarySigner *NotarySignerComponent `json:"notarySigner"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}

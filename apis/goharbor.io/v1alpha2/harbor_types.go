package v1alpha2

import (
	"fmt"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
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

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
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
	InternalTLS HarborInternalTLSSpec `json:"internalTLS"`

	// +kubebuilder:validation:Optional
	Persistence HarborPersistenceSpec `json:"persistence,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	LogLevel harbormetav1.HarborLogLevel `json:"logLevel,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	HarborAdminPasswordRef string `json:"harborAdminPasswordRef"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
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
	Portal harbormetav1.ComponentSpec `json:"portal,omitempty"`

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
	Notary *NotaryComponentSpec `json:"notary,omitempty"`

	// +kubebuilder:validation:Required
	Redis ExternalRedisSpec `json:"redis"`

	// +kubebuilder:validation:Required
	Database HarborDatabaseSpec `json:"database"`
}

type HarborDatabaseSpec struct {
	harbormetav1.PostgresCredentials `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Hosts []harbormetav1.PostgresHostSpec `json:"hosts"`

	// +kubebuilder:validation:Optional
	SSLMode harbormetav1.PostgresSSLMode `json:"sslMode,omitempty"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix,omitempty"`
}

func (r *HarborDatabaseSpec) GetPostgresqlConnection(component harbormetav1.Component) (*harbormetav1.PostgresConnectionWithParameters, error) {
	sslMode := r.SSLMode

	var databaseName string

	switch component { // nolint:exhaustive
	case harbormetav1.CoreComponent:
		databaseName = harbormetav1.CoreDatabase
	case harbormetav1.NotarySignerComponent:
		sslMode = r.getSSLModeForNotary()
		databaseName = harbormetav1.NotarySignerDatabase
	case harbormetav1.NotaryServerComponent:
		sslMode = r.getSSLModeForNotary()
		databaseName = harbormetav1.NotaryServerDatabase
	case harbormetav1.ClairComponent:
		databaseName = harbormetav1.ClairDatabase
	default:
		return nil, harbormetav1.ErrUnsupportedComponent
	}

	return &harbormetav1.PostgresConnectionWithParameters{
		PostgresConnection: harbormetav1.PostgresConnection{
			PostgresCredentials: r.PostgresCredentials,
			Database:            r.Prefix + databaseName,
			Hosts:               r.Hosts,
		},
		Parameters: map[string]string{
			harbormetav1.PostgresSSLModeKey: string(sslMode),
		},
	}, nil
}

func (r *HarborDatabaseSpec) getSSLModeForNotary() harbormetav1.PostgresSSLMode {
	switch r.SSLMode { //nolint:exhaustive
	case harbormetav1.PostgresSSLModeAllow:
		return harbormetav1.PostgresSSLModePrefer
	default:
		return r.SSLMode
	}
}

type NotaryComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`
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
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	PasswordRef string `json:"passwordRef,omitempty"`
}

func (r *HarborComponentsSpec) RedisDSN(component harbormetav1.ComponentWithRedis) OpacifiedDSN {
	return OpacifiedDSN{
		DSN:         fmt.Sprintf("redis://%s:%d/%d", r.Redis.Address, r.Redis.Port, component.Index()),
		PasswordRef: r.Redis.PasswordRef,
	}
}

type CoreComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Required
	TokenIssuer cmmeta.ObjectReference `json:"tokenIssuer,omitempty"`
}

type JobServiceComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10
	WorkerCount int32 `json:"workerCount,omitempty"`
}

type RegistryComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	RelativeURLs *bool `json:"relativeURLs,omitempty"`

	// +kubebuilder:validation:Optional
	StorageMiddlewares []RegistryMiddlewareSpec `json:"storageMiddlewares,omitempty"`
}

type ChartMuseumComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`
}

type ClairComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// One of clair redis dsn or global redis component must be specified
	Redis *OpacifiedDSN `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="12h"
	Interval *metav1.Duration `json:"updatersInterval,omitempty"`
}

type TrivyComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`
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
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`
}

func (r *HarborInternalTLSSpec) IsEnabled() bool {
	return r != nil && r.Enabled
}

const CertificateAuthoritySecretConfigKey = "certificate-authority-secret"

func (r *HarborInternalTLSSpec) GetScheme() string {
	if !r.IsEnabled() {
		return "http"
	}

	return "https"
}

type ErrUnsupportedComponent harbormetav1.ComponentWithTLS

func (err ErrUnsupportedComponent) Error() string {
	return fmt.Sprintf("%s is not supported", string(err))
}

func (r *HarborInternalTLSSpec) GetInternalPort(component harbormetav1.ComponentWithTLS) (int32, error) {
	if !r.IsEnabled() {
		return harbormetav1.HTTPPort, nil
	}

	return harbormetav1.HTTPSPort, nil
}

func (r *HarborInternalTLSSpec) GetComponentTLSSpec(certificateRef string) *harbormetav1.ComponentsTLSSpec {
	if !r.IsEnabled() {
		return nil
	}

	return &harbormetav1.ComponentsTLSSpec{
		CertificateRef: certificateRef,
	}
}

type HarborExposeSpec struct {
	// +kubebuilder:validation:Required
	Core HarborExposeComponentSpec `json:"core"`

	// +kubebuilder:validation:Optional
	Notary *HarborExposeComponentSpec `json:"notary,omitempty"`
}

type HarborExposeComponentSpec struct {
	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	Ingress *HarborExposeIngressSpec `json:"ingress,omitempty"`

	// TODO Add supports to ClusterIP, LoadBalancer and NodePort by deploying the nginx component
}

type HarborExposeIngressSpec struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Enum={"gce","ncp","default"}
	// Set to the type of ingress controller.
	Controller harbormetav1.IngressController `json:"controller,omitempty"`

	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}

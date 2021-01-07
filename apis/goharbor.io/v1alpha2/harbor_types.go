package v1alpha2

import (
	"fmt"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
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
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.externalURL`,description="The public URL to the Harbor application",priority=5
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
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
	HarborComponentsSpec `json:",inline"`

	ImageSource *ImageSourceSpec `json:"imageSource,omitempty"`

	// +kubebuilder:validation:Required
	Expose HarborExposeSpec `json:"expose"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.*"
	ExternalURL string `json:"externalURL"`

	// +kubebuilder:validation:Optional
	InternalTLS HarborInternalTLSSpec `json:"internalTLS"`

	// Skip OpenAPI schema validation
	// Use validating webhook to do verification (field required)
	// +kubebuilder:validation:Optional
	ImageChartStorage *HarborStorageImageChartStorageSpec `json:"imageChartStorage,omitempty"`

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
	// +kubebuilder:default="RollingUpdate"
	UpdateStrategyType appsv1.DeploymentStrategyType `json:"updateStrategyType,omitempty"`

	// +kubebuilder:validation:Optional
	Proxy *CoreProxySpec `json:"proxy,omitempty"`

	// +kubebuilder:validation:Optional
	// The version of the harbor, eg 2.1.2
	Version string `json:"version,omitempty"`
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
	Trivy *TrivyComponentSpec `json:"trivy,omitempty"`

	// +kubebuilder:validation:Optional
	Notary *NotaryComponentSpec `json:"notary,omitempty"`

	// Skip OpenAPI schema validation
	// Use validating webhook to do verification (field required)
	// +kubebuilder:validation:Optional
	Redis *ExternalRedisSpec `json:"redis,omitempty"`

	// Skip OpenAPI schema validation
	// Use validating webhook to do verification (field required)
	// +kubebuilder:validation:Optional
	Database *HarborDatabaseSpec `json:"database,omitempty"`
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
	// +kubebuilder:validation:Optional
	Server harbormetav1.ComponentSpec `json:"server"`

	// +kubebuilder:validation:Optional
	Signer harbormetav1.ComponentSpec `json:"signer"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// Inject migration configuration to notary resources
	MigrationEnabled *bool `json:"migrationEnabled,omitempty"`
}

func (r *NotaryComponentSpec) IsMigrationEnabled() bool {
	return r != nil && (r.MigrationEnabled == nil || *r.MigrationEnabled)
}

type ExternalRedisSpec struct {
	harbormetav1.RedisHostSpec    `json:",inline"`
	harbormetav1.RedisCredentials `json:",inline"`
}

func (r *HarborComponentsSpec) RedisConnection(component harbormetav1.ComponentWithRedis) harbormetav1.RedisConnection {
	return harbormetav1.RedisConnection{
		RedisCredentials: r.Redis.RedisCredentials,
		RedisHostSpec:    r.Redis.RedisHostSpec,
		Database:         component.Index(),
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

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Harbor defaults ChartMuseum to returning relative urls,
	// if you want using absolute url you should enable it
	AbsoluteURL bool `json:"absoluteUrl"`
}

type TrivyComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// The name of the secret containing the token to connect to GitHub API.
	GithubTokenRef string `json:"githubTokenRef,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to enable or disable Trivy DB downloads from GitHub
	SkipUpdate bool `json:"skipUpdate"`

	// +kubebuilder:validation:Required
	Storage HarborStorageTrivyStorageSpec `json:"storage"`
}

type HarborStorageImageChartStorageSpec struct {
	// +kubebuilder:validation:Optional
	Redirect RegistryStorageRedirectSpec `json:"redirect,omitempty"`

	// +kubebuilder:validation:Optional
	// FileSystem is an implementation of the storagedriver.StorageDriver interface which uses the local filesystem.
	// The local filesystem can be a remote volume.
	// See: https://docs.docker.com/registry/storage-drivers/filesystem/
	FileSystem *HarborStorageImageChartStorageFileSystemSpec `json:"filesystem,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Amazon S3 or S3 compatible services for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/s3/
	S3 *HarborStorageImageChartStorageS3Spec `json:"s3,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface that uses OpenStack Swift for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/swift/
	Swift *HarborStorageImageChartStorageSwiftSpec `json:"swift,omitempty"`
}

type HarborStorageTrivyStorageSpec struct {
	// +kubebuilder:validation:Optional
	// ReportsPersistentVolume specify the persistent volume used to store Trivy reports.
	// If empty, empty dir will be used.
	ReportsPersistentVolume *HarborStoragePersistentVolumeSpec `json:"reportsPersistentVolume,omitempty"`

	// +kubebuilder:validation:Optional
	// CachePersistentVolume specify the persistent volume used to store Trivy cache.
	// If empty, empty dir will be used.
	CachePersistentVolume *HarborStoragePersistentVolumeSpec `json:"cachePersistentVolume,omitempty"`
}

const (
	S3DriverName         = "s3"
	SwiftDriverName      = "swift"
	FileSystemDriverName = "filesystem"
)

func (r *HarborStorageImageChartStorageSpec) ProviderName() string {
	if r.S3 != nil {
		return S3DriverName
	}

	if r.Swift != nil {
		return SwiftDriverName
	}

	return FileSystemDriverName
}

func (r *HarborStorageImageChartStorageSpec) Validate() error {
	if r == nil {
		return ErrNoStorageConfiguration
	}

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

type HarborStorageImageChartStorageFileSystemSpec struct {
	// +kubebuilder:validation:Optional
	ChartPersistentVolume *HarborStoragePersistentVolumeSpec `json:"chartPersistentVolume,omitempty"`

	// +kubebuilder:validation:Required
	RegistryPersistentVolume HarborStorageRegistryPersistentVolumeSpec `json:"registryPersistentVolume"`
}

type HarborStorageRegistryPersistentVolumeSpec struct {
	HarborStoragePersistentVolumeSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=25
	// +kubebuilder:default=100
	MaxThreads int32 `json:"maxthreads,omitempty"`
}

type HarborStorageImageChartStorageS3Spec struct {
	RegistryStorageDriverS3Spec `json:",inline"`
}

func (r *HarborStorageImageChartStorageS3Spec) ChartMuseum() *ChartMuseumChartStorageDriverAmazonSpec {
	return &ChartMuseumChartStorageDriverAmazonSpec{
		AccessKeyID:     r.AccessKey,
		AccessSecretRef: r.SecretKeyRef,
		Bucket:          r.Bucket,
		Endpoint:        r.RegionEndpoint,
		Prefix:          r.RootDirectory,
		Region:          r.Region,
	}
}

func (r *HarborStorageImageChartStorageS3Spec) Registry() *RegistryStorageDriverS3Spec {
	return &r.RegistryStorageDriverS3Spec
}

type HarborStorageImageChartStorageSwiftSpec struct {
	RegistryStorageDriverSwiftSpec `json:",inline"`
}

func (r *HarborStorageImageChartStorageSwiftSpec) ChartMuseum() *ChartMuseumChartStorageDriverOpenStackSpec {
	return &ChartMuseumChartStorageDriverOpenStackSpec{
		AuthenticationURL: r.AuthURL,
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

func (r *HarborStorageImageChartStorageSwiftSpec) Registry() *RegistryStorageDriverSwiftSpec {
	return &r.RegistryStorageDriverSwiftSpec
}

type HarborStoragePersistentVolumeSpec struct {
	corev1.PersistentVolumeClaimVolumeSource `json:",inline"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix,omitempty"`
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
	return fmt.Sprintf("%s is not supported", string(rune(err)))
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

type ImageSourceSpec struct {
	// +kubebuilder:validation:Required
	// The default repository for the images of the components. eg docker.io/goharbor/
	Repository string `json:"repository,omitempty"`

	// +kubebuilder:validation:Optional
	// The tag suffix for the images of the images of the components. eg '-patch1'
	TagSuffix string `json:"tagSuffix,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"Always","Never","IfNotPresent"}
	// Image pull policy.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
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

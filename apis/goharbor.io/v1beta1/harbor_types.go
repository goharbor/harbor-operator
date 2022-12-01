package v1beta1

import (
	"context"
	"fmt"
	"path"
	"strings"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/image"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The version to the Harbor application",priority=5
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// Harbor is the Schema for the harbors API.
type Harbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HarborSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

func (h *Harbor) GetComponentSpec(ctx context.Context, component harbormetav1.Component) harbormetav1.ComponentSpec {
	var spec harbormetav1.ComponentSpec

	h.deepCopyComponentSpecInto(ctx, component, &spec)
	h.deepCopyImageSpecInto(ctx, component, &spec)

	return spec
}

func (h *Harbor) deepCopyComponentSpecInto(_ context.Context, component harbormetav1.Component, spec *harbormetav1.ComponentSpec) {
	switch component {
	case harbormetav1.ChartMuseumComponent:
		if h.Spec.ChartMuseum != nil {
			h.Spec.ChartMuseum.ComponentSpec.DeepCopyInto(spec)
		}
	case harbormetav1.CoreComponent:
		h.Spec.Core.ComponentSpec.DeepCopyInto(spec)
	case harbormetav1.ExporterComponent:
		h.Spec.Exporter.ComponentSpec.DeepCopyInto(spec)
	case harbormetav1.JobServiceComponent:
		h.Spec.JobService.ComponentSpec.DeepCopyInto(spec)
	case harbormetav1.NotaryServerComponent:
		if h.Spec.Notary != nil {
			h.Spec.Notary.Server.DeepCopyInto(spec)
		}
	case harbormetav1.NotarySignerComponent:
		if h.Spec.Notary != nil {
			h.Spec.Notary.Signer.DeepCopyInto(spec)
		}
	case harbormetav1.PortalComponent:
		h.Spec.Portal.ComponentSpec.DeepCopyInto(spec)
	case harbormetav1.RegistryComponent:
		h.Spec.Registry.ComponentSpec.DeepCopyInto(spec)
	case harbormetav1.RegistryControllerComponent:
		if h.Spec.RegistryController != nil {
			h.Spec.RegistryController.DeepCopyInto(spec)
		}

		h.deepCopyNodeSelectorAndTolerationsOfRegistryInto(spec)
	case harbormetav1.TrivyComponent:
		if h.Spec.Trivy != nil {
			h.Spec.Trivy.ComponentSpec.DeepCopyInto(spec)
		}
	}
}

func (h *Harbor) deepCopyNodeSelectorAndTolerationsOfRegistryInto(spec *harbormetav1.ComponentSpec) {
	if h.Spec.ImageChartStorage.FileSystem == nil {
		return
	}

	if len(spec.NodeSelector) == 0 && len(h.Spec.Registry.NodeSelector) != 0 {
		in, out := &h.Spec.Registry.NodeSelector, &spec.NodeSelector
		*out = make(map[string]string, len(*in))

		for key, val := range *in {
			(*out)[key] = val
		}
	}

	if len(spec.Tolerations) == 0 && len(h.Spec.Registry.Tolerations) != 0 {
		in, out := &h.Spec.Registry.Tolerations, &spec.Tolerations
		*out = make([]corev1.Toleration, len(*in))

		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (h *Harbor) GetComponentProxySpec(component harbormetav1.Component) *harbormetav1.ProxySpec {
	if h.Spec.Proxy == nil {
		return nil
	}

	for _, c := range h.Spec.Proxy.Components {
		if c == component.String() {
			return &h.Spec.Proxy.ProxySpec
		}
	}

	return nil
}

func (h *Harbor) deepCopyImageSpecInto(ctx context.Context, component harbormetav1.Component, spec *harbormetav1.ComponentSpec) {
	imageSource := h.Spec.ImageSource
	if imageSource == nil {
		return
	}

	if spec.Image == "" && (imageSource.Repository != "" || imageSource.TagSuffix != "") {
		getImageOptions := []image.Option{
			image.WithRepository(imageSource.Repository),
			image.WithTagSuffix(imageSource.TagSuffix),
			image.WithHarborVersion(h.Spec.Version),
		}
		spec.Image, _ = image.GetImage(ctx, component.String(), getImageOptions...)
	}

	if spec.ImagePullPolicy == nil && imageSource.ImagePullPolicy != nil {
		in, out := &imageSource.ImagePullPolicy, &spec.ImagePullPolicy
		*out = new(corev1.PullPolicy)
		**out = **in
	}

	if len(spec.ImagePullSecrets) == 0 && len(imageSource.ImagePullSecrets) != 0 {
		in, out := &imageSource.ImagePullSecrets, &spec.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
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

	ImageSource *harbormetav1.ImageSourceSpec `json:"imageSource,omitempty"`

	// +kubebuilder:validation:Required
	Expose HarborExposeSpec `json:"expose"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.*"
	ExternalURL string `json:"externalURL"`

	// +kubebuilder:validation:Optional
	InternalTLS HarborInternalTLSSpec `json:"internalTLS"`

	// +kubebuilder:validation:Required
	ImageChartStorage *HarborStorageImageChartStorageSpec `json:"imageChartStorage"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	LogLevel harbormetav1.HarborLogLevel `json:"logLevel,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	HarborAdminPasswordRef string `json:"harborAdminPasswordRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="RollingUpdate"
	UpdateStrategyType appsv1.DeploymentStrategyType `json:"updateStrategyType,omitempty"`

	// +kubebuilder:validation:Optional
	Proxy *HarborProxySpec `json:"proxy,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[0-9]+\\.[0-9]+\\.[0-9]+"
	// The version of the harbor, eg 2.1.2
	Version string `json:"version"`

	// +kubebuilder:validation:Optional
	// Network settings for the harbor
	Network *harbormetav1.Network `json:"network,omitempty"`

	// +kubebuilder:validation:Optional
	// Trace settings for the harbor
	Trace *harbormetav1.TraceSpec `json:"trace,omitempty"`
}

func (spec *HarborSpec) ValidateNotary() *field.Error {
	return nil
}

func (spec *HarborSpec) ValidateRegistryController() *field.Error {
	if spec.RegistryController == nil {
		return nil
	}

	// nodeSelector and tolerations must equal with registry's
	// when the image chart storage is filesystem and the access mode of pvc is ReadWriteOnce
	// TODO: check the access mode of pvc is ReadWriteOnce
	if spec.ImageChartStorage != nil && spec.ImageChartStorage.FileSystem != nil {
		if len(spec.RegistryController.NodeSelector) > 0 &&
			!equality.Semantic.DeepEqual(spec.RegistryController.NodeSelector, spec.Registry.NodeSelector) {
			p := field.NewPath("spec").Child("registryctl", "nodeSelector")

			return field.Forbidden(p, "must be empty or equal with spec.registry.nodeSelector")
		}

		if len(spec.RegistryController.Tolerations) > 0 &&
			!equality.Semantic.DeepEqual(spec.RegistryController.Tolerations, spec.Registry.Tolerations) {
			p := field.NewPath("spec").Child("registryctl", "tolerations")

			return field.Forbidden(p, "must be empty or equal with spec.registry.tolerations")
		}
	}

	return nil
}

type HarborComponentsSpec struct {
	// +kubebuilder:validation:Optional
	Portal *PortalComponentSpec `json:"portal,omitempty"`

	// +kubebuilder:validation:Required
	Core CoreComponentSpec `json:"core,omitempty"`

	// +kubebuilder:validation:Required
	JobService JobServiceComponentSpec `json:"jobservice,omitempty"`

	// +kubebuilder:validation:Required
	Registry RegistryComponentSpec `json:"registry,omitempty"`

	// +kubebuilder:validation:Optional
	RegistryController *harbormetav1.ComponentSpec `json:"registryctl,omitempty"`

	// +kubebuilder:validation:Optional
	ChartMuseum *ChartMuseumComponentSpec `json:"chartmuseum,omitempty"`

	// +kubebuilder:validation:Optional
	Exporter *ExporterComponentSpec `json:"exporter,omitempty"`

	// +kubebuilder:validation:Optional
	Trivy *TrivyComponentSpec `json:"trivy,omitempty"`

	// +kubebuilder:validation:Optional
	Notary *NotaryComponentSpec `json:"notary,omitempty"`

	// +kubebuilder:validation:Required
	Redis *ExternalRedisSpec `json:"redis"`

	Database *HarborDatabaseSpec `json:"database"`
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

	switch component { //nolint:exhaustive
	case harbormetav1.CoreComponent:
		databaseName = harbormetav1.CoreDatabase
	case harbormetav1.ExporterComponent:
		// exporter requires to access the database of core component
		databaseName = harbormetav1.CoreDatabase
	case harbormetav1.NotarySignerComponent:
		sslMode = r.getSSLModeForNotary()
		databaseName = harbormetav1.NotarySignerDatabase
	case harbormetav1.NotaryServerComponent:
		sslMode = r.getSSLModeForNotary()
		databaseName = harbormetav1.NotaryServerDatabase
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

type PortalComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`
}

type CoreComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Required
	TokenIssuer cmmeta.ObjectReference `json:"tokenIssuer"`

	// +kubebuilder:validation:Optional
	Metrics *harbormetav1.MetricsSpec `json:"metrics,omitempty"`
}

type JobServiceComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10
	WorkerCount int32 `json:"workerCount,omitempty"`

	// +kubebuilder:validation:Optional
	Metrics *harbormetav1.MetricsSpec `json:"metrics,omitempty"`

	// +kubebuilder:validation:Optional
	Storage *HarborStorageJobServiceStorageSpec `json:"storage,omitempty"`
}

type RegistryComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	RelativeURLs *bool `json:"relativeURLs,omitempty"`

	// +kubebuilder:validation:Optional
	StorageMiddlewares []RegistryMiddlewareSpec `json:"storageMiddlewares,omitempty"`

	// +kubebuilder:validation:Optional
	Metrics *harbormetav1.MetricsSpec `json:"metrics,omitempty"`
}

type ChartMuseumComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Harbor defaults ChartMuseum to returning relative urls,
	// if you want using absolute url you should enable it
	AbsoluteURL bool `json:"absoluteUrl"`
}

type ExporterComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	Cache HarborExporterCacheSpec `json:"cache,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=8001
	// +kubebuilder:validation:Minimum=1
	// The port of the exporter.
	Port int32 `json:"port"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/metrics"
	// +kubebuilder:validation:Pattern="/.+"
	// The metrics path of the exporter.
	Path string `json:"path"`
}

type HarborExporterCacheSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?"
	// +kubebuilder:default="30s"
	// The duration to cache info from the database and core.
	Duration *metav1.Duration `json:"duration,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?"
	// +kubebuilder:default="4h"
	// The interval to clean the cache info from the database and core.
	CleanInterval *metav1.Duration `json:"cleanInterval,omitempty"`
}

type TrivyComponentSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	// The name of the secret containing the token to connect to GitHub API.
	GithubTokenRef string `json:"githubTokenRef,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to enable or disable Trivy DB downloads from GitHub
	SkipUpdate bool `json:"skipUpdate"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Option prevents Trivy from sending API requests to identify dependencies.
	// This option doesn’t affect DB download. You need to specify "skip-update" as well as "offline-scan" in an air-gapped environment.
	OfflineScan bool `json:"offlineScan"`

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

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Microsoft Azure Blob Storage for object storage.
	// See https://docs.docker.com/registry/storage-drivers/azure/
	Azure *HarborStorageImageChartStorageAzureSpec `json:"azure,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Google Cloud for object storage.
	// See https://docs.docker.com/registry/storage-drivers/gcs/
	Gcs *HarborStorageImageChartStorageGcsSpec `json:"gcs,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Alibaba Cloud for object storage.
	// See https://docs.docker.com/registry/storage-drivers/oss/
	Oss *HarborStorageImageChartStorageOssSpec `json:"oss,omitempty"`
}

type HarborStorageJobServiceStorageSpec struct {
	// +kubebuilder:validation:Optional
	// ScanDataExportsPersistentVolume specify the persistent volume used to store data exports.
	// If empty, empty dir will be used.
	ScanDataExportsPersistentVolume *HarborStoragePersistentVolumeSpec `json:"scanDataExportsPersistentVolume,omitempty"`
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
	AzureDriverName      = "azure"
	GcsDriverName        = "gcs"
	OssDriverName        = "oss"
)

func (r *HarborStorageImageChartStorageSpec) ProviderName() string {
	if r.S3 != nil {
		return S3DriverName
	}

	if r.Swift != nil {
		return SwiftDriverName
	}

	if r.Azure != nil {
		return AzureDriverName
	}

	if r.Gcs != nil {
		return GcsDriverName
	}

	if r.Oss != nil {
		return OssDriverName
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

	if r.Azure != nil {
		found++
	}

	if r.Gcs != nil {
		found++
	}

	if r.Oss != nil {
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

type HarborStorageImageChartStorageAzureSpec struct {
	RegistryStorageDriverAzureSpec `json:",inline"`
}

type HarborStorageImageChartStorageOssSpec struct {
	RegistryStorageDriverOssSpec `json:",inline"`
}

func (r *HarborStorageImageChartStorageOssSpec) ChartMuseum() *ChartMuseumChartStorageDriverOssSpec {
	return &ChartMuseumChartStorageDriverOssSpec{
		Endpoint:        r.getEndpoint(),
		AccessKeyID:     r.AccessKeyID,
		AccessSecretRef: r.AccessSecretRef,
		Bucket:          r.Bucket,
		PathPrefix:      r.PathPrefix,
	}
}

func (r *HarborStorageImageChartStorageOssSpec) Registry() *RegistryStorageDriverOssSpec {
	return &r.RegistryStorageDriverOssSpec
}

func (r *HarborStorageImageChartStorageOssSpec) getEndpoint() string {
	if r.Endpoint != "" {
		return r.Endpoint
	}

	if r.Internal {
		return fmt.Sprintf("%s-internal.aliyuncs.com", r.Region)
	}

	return fmt.Sprintf("%s.aliyuncs.com", r.Region)
}

type HarborStorageImageChartStorageGcsSpec struct {
	RegistryStorageDriverGcsSpec `json:",inline"`
}

func (r *HarborStorageImageChartStorageGcsSpec) ChartMuseum() *ChartMuseumChartStorageDriverGcsSpec {
	return &ChartMuseumChartStorageDriverGcsSpec{
		KeyDataSecretRef: r.KeyDataRef,
		Bucket:           r.Bucket,
		PathPrefix:       r.PathPrefix,
		ChunkSize:        r.ChunkSize,
	}
}

func (r *HarborStorageImageChartStorageGcsSpec) Registry() *RegistryStorageDriverGcsSpec {
	return &r.RegistryStorageDriverGcsSpec
}

func (r *HarborStorageImageChartStorageAzureSpec) ChartMuseum() *ChartMuseumChartStorageDriverAzureSpec {
	return &ChartMuseumChartStorageDriverAzureSpec{
		AccountName:   r.AccountName,
		AccountKeyRef: r.AccountKeyRef,
		Container:     r.Container,
		BaseURL:       r.BaseURL,
		PathPrefix:    r.PathPrefix,
	}
}

func (r *HarborStorageImageChartStorageAzureSpec) Registry() *RegistryStorageDriverAzureSpec {
	return &r.RegistryStorageDriverAzureSpec
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

func (r *HarborInternalTLSSpec) GetScheme() string {
	if !r.IsEnabled() {
		return "http"
	}

	return "https"
}

type ErrUnsupportedComponent harbormetav1.ComponentWithTLS

func (err ErrUnsupportedComponent) Error() string {
	return fmt.Sprintf("%s is not supported", harbormetav1.ComponentWithTLS(err).String())
}

func (r *HarborInternalTLSSpec) GetInternalPort(component harbormetav1.ComponentWithTLS) int32 {
	if !r.IsEnabled() {
		return harbormetav1.HTTPPort
	}

	return harbormetav1.HTTPSPort
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
	// The ingress of the notary, required when notary component enabled.
	Notary *HarborExposeComponentSpec `json:"notary,omitempty"`
}

type HarborExposeComponentSpec struct {
	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	Ingress *HarborExposeIngressSpec `json:"ingress,omitempty"`
}

type HarborExposeIngressSpec struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	// Set to the type of ingress controller.
	Controller harbormetav1.IngressController `json:"controller,omitempty"`

	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +kubebuilder:validation:Optional
	IngressClassName *string `json:"ingressClassName,omitempty"`
}

// CertificateInjection defines the certs injection.
type CertificateInjection struct {
	// +kubebuilder:validation:Optional
	CertificateRefs []string `json:"certificateRefs,omitempty"`
}

// ShouldInject returns whether should inject certs.
func (ci CertificateInjection) ShouldInject() bool {
	return len(ci.CertificateRefs) > 0
}

// GenerateVolumes generates volumes.
func (ci CertificateInjection) GenerateVolumes() []corev1.Volume {
	volumes := make([]corev1.Volume, 0, len(ci.CertificateRefs))
	for _, ref := range ci.CertificateRefs {
		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("%s-certifacts", ref),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: ref,
				},
			},
		})
	}

	return volumes
}

// GenerateVolumeMounts generates volumeMounts.
func (ci CertificateInjection) GenerateVolumeMounts() []corev1.VolumeMount {
	volumeMounts := make([]corev1.VolumeMount, 0, len(ci.CertificateRefs))
	for _, ref := range ci.CertificateRefs {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-certifacts", ref),
			MountPath: path.Join("/harbor_cust_cert", fmt.Sprintf("%s.crt", ref)),
			SubPath:   strings.TrimLeft(corev1.ServiceAccountRootCAKey, "/"),
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

type HarborProxySpec struct {
	harbormetav1.ProxySpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={core,jobservice,trivy}
	Components []string `json:"components,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}

package v1beta1

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	KindDatabaseZlandoPostgreSQL = "Zlando/PostgreSQL"
	KindDatabasePostgreSQL       = "PostgreSQL"
	KindStorageOss               = "Oss"
	KindStorageGcs               = "Gcs"
	KindStorageAzure             = "Azure"
	KindStorageMinIO             = "MinIO"
	KindStorageSwift             = "Swift"
	KindStorageS3                = "S3"
	KindStorageFileSystem        = "FileSystem"
	KindCacheRedisFailover       = "RedisFailover"
	KindCacheRedis               = "Redis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var HarborClusterGVK = schema.GroupVersionKind{
	Group:   GroupVersion.Group,
	Version: GroupVersion.Version,
	Kind:    "HarborCluster",
}

// HarborClusterSpec defines the desired state of HarborCluster.
type HarborClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	EmbeddedHarborSpec `json:",inline"`

	// Cache configuration for in-cluster cache services
	// +kubebuilder:validation:Required
	Cache Cache `json:"cache"`

	// Database configuration for in-cluster database service
	// +kubebuilder:validation:Required
	Database Database `json:"database"`

	// Storage configuration for in-cluster storage service
	// +kubebuilder:validation:Required
	Storage Storage `json:"storage"`

	// +kubebuilder:validation:Optional
	// Network settings for the harbor
	Network *harbormetav1.Network `json:"network,omitempty"`

	// +kubebuilder:validation:Optional
	// Trace settings for the harbor
	Trace *harbormetav1.TraceSpec `json:"trace,omitempty"`
}

type EmbeddedHarborSpec struct {
	EmbeddedHarborComponentsSpec `json:",inline"`

	ImageSource *harbormetav1.ImageSourceSpec `json:"imageSource,omitempty"`

	// +kubebuilder:validation:Required
	Expose HarborExposeSpec `json:"expose"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.*"
	ExternalURL string `json:"externalURL"`

	// +kubebuilder:validation:Optional
	InternalTLS HarborInternalTLSSpec `json:"internalTLS"`

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
}

type EmbeddedHarborComponentsSpec struct {
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
}

type Cache struct {
	// Set the kind of cache service to be used. Only support Redis now.
	// +kubebuilder:validation:Enum={Redis,RedisFailover}
	Kind string `json:"kind"`

	// RedisSpec is the specification of redis.
	// +kubebuilder:validation:Required
	Spec *CacheSpec `json:"spec"`
}

type CacheSpec struct {
	// +kubebuilder:validation:Optional
	Redis *ExternalRedisSpec `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	RedisFailover *RedisFailoverSpec `json:"redisFailover,omitempty"`
}

type RedisFailoverSpec struct {
	harbormetav1.ImageSpec `json:",inline"`

	OperatorVersion string `json:"operatorVersion"`

	// +kubebuilder:validation:Optional
	// Server is the configuration of the redis server.
	Server *RedisServer `json:"server,omitempty"`
	// +kubebuilder:validation:Optional
	// Sentinel is the configuration of the redis sentinel.
	Sentinel *RedisSentinel `json:"sentinel,omitempty"`
}

type RedisSentinel struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// Replicas is the instance number of redis sentinel.
	Replicas int `json:"replicas,omitempty"`
}

type RedisServer struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// Replicas is the instance number of redis server.
	Replicas int `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	// Resources is the resources requests and limits for redis.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Optional
	// StorageClassName is the storage class name of the redis storage.
	StorageClassName string `json:"storageClassName,omitempty"`

	// +kubebuilder:validation:Optional
	// Storage is the size of the redis storage.
	Storage string `json:"storage,omitempty"`
}

type Database struct {
	// Set the kind of which database service to be used, Only support PostgreSQL now.
	// +kubebuilder:validation:Enum={PostgreSQL,Zlando/PostgreSQL}
	Kind string `json:"kind"`

	// +kubebuilder:validation:Required
	Spec DatabaseSpec `json:"spec"`
}

type DatabaseSpec struct {
	// +kubebuilder:validation:Optional
	PostgreSQL *PostgreSQLSpec `json:"postgresql,omitempty"`

	// ZlandoPostgreSQL
	ZlandoPostgreSQL *ZlandoPostgreSQLSpec `json:"zlandoPostgreSql,omitempty"`
}

type PostgreSQLSpec struct {
	HarborDatabaseSpec `json:",inline"`
}

type ZlandoPostgreSQLSpec struct {
	harbormetav1.ImageSpec `json:",inline"`

	OperatorVersion string `json:"operatorVersion"`

	Storage          string                      `json:"storage,omitempty"`
	Replicas         int                         `json:"replicas,omitempty"`
	StorageClassName string                      `json:"storageClassName,omitempty"`
	Resources        corev1.ResourceRequirements `json:"resources,omitempty"`
	SslConfig        string                      `json:"sslConfig,omitempty"`
	ConnectTimeout   int                         `json:"connectTimeout,omitempty"`
}

type Storage struct {
	// Kind of which storage service to be used. Only support MinIO now.
	// +kubebuilder:validation:Enum={MinIO,S3,Swift,FileSystem,Azure,Gcs,Oss}
	Kind string `json:"kind"`

	Spec StorageSpec `json:"spec"`
}

// the spec of Storage.
type StorageSpec struct {
	// inCluster options.
	// +kubebuilder:validation:Optional
	MinIO *MinIOSpec `json:"minIO,omitempty"`
	// +kubebuilder:validation:Optional
	FileSystem *FileSystemSpec `json:"fileSystem,omitempty"`
	// +kubebuilder:validation:Optional
	S3 *S3Spec `json:"s3,omitempty"`
	// +kubebuilder:validation:Optional
	Swift *SwiftSpec `json:"swift,omitempty"`
	// +kubebuilder:validation:Optional
	Azure *AzureSpec `json:"azure,omitempty"`
	// +kubebuilder:validation:Optional
	Gcs *GcsSpec `json:"gcs,omitempty"`
	// +kubebuilder:validation:Optional
	Oss *OssSpec `json:"oss,omitempty"`
	// Determine if the redirection of minio storage is disabled.
	// +kubebuilder:validation:Optional
	Redirect *StorageRedirectSpec `json:"redirect,omitempty"`
}

// StorageRedirectSpec defines if the redirection is disabled.
type StorageRedirectSpec struct {
	// Default is true
	// +kubebuilder:default:=true
	Enable bool `json:"enable"`
	// +kubebuilder:validation:Optional
	Expose *HarborExposeComponentSpec `json:"expose,omitempty"`
}

type FileSystemSpec struct {
	HarborStorageImageChartStorageFileSystemSpec `json:",inline"`
}

type S3Spec struct {
	HarborStorageImageChartStorageS3Spec `json:",inline"`
}

type AzureSpec struct {
	HarborStorageImageChartStorageAzureSpec `json:",inline"`
}

type GcsSpec struct {
	HarborStorageImageChartStorageGcsSpec `json:",inline"`
}

type OssSpec struct {
	HarborStorageImageChartStorageOssSpec `json:",inline"`
}

type SwiftSpec struct {
	HarborStorageImageChartStorageSwiftSpec `json:",inline"`
}

type MinIOSpec struct {
	harbormetav1.ImageSpec `json:",inline"`

	// the version of minIO operator
	// +kubebuilder:default:="4.0.6"
	OperatorVersion string `json:"operatorVersion"`

	// deprecated Determine if the redirection of minio storage is disabled.
	// +kubebuilder:validation:Optional
	Redirect *StorageRedirectSpec `json:"redirect,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Reference to the secret containing the MinIO access key and secret key.
	SecretRef string `json:"secretRef,omitempty"`
	// Supply number of replicas.
	// For standalone mode, supply 1. For distributed mode, supply 4 to 16 drives (should be even).
	// Note that the operator does not support upgrading from standalone to distributed mode.
	// +kubebuilder:validation:Minimum:=1
	Replicas int32 `json:"replicas"`
	// Number of persistent volumes that will be attached per server
	// +kubebuilder:validation:Minimum:=1
	VolumesPerServer int32 `json:"volumesPerServer"`
	// VolumeClaimTemplate allows a user to specify how volumes inside a MinIOInstance
	// +kubebuilder:validation:Optional
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +kubebuilder:validation:Optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// MinIOClientSpec the spec for the mc
	// +kubebuilder:validation:Optional
	MinIOClientSpec *MinIOClientSpec `json:"mc,omitempty"`
}

func (spec *MinIOSpec) GetMinIOClientImage() string {
	if spec.MinIOClientSpec == nil {
		return ""
	}

	return spec.MinIOClientSpec.Image
}

type MinIOClientSpec struct {
	harbormetav1.ImageSpec `json:",inline"`
}

// HarborClusterStatus defines the observed state of HarborCluster.
type HarborClusterStatus struct {
	// +kubebuilder:validation:Optional
	Operator harbormetav1.OperatorStatus `json:"operator,omitempty"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Status indicates the overall status of the Harbor cluster
	// Status can be "unknown", "creating", "healthy" and "unhealthy"
	Status ClusterStatus `json:"status"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Revision of the status
	// Use unix nano
	Revision int64 `json:"revision"`

	// Conditions of each components
	Conditions []HarborClusterCondition `json:"conditions,omitempty"`
}

// ClusterStatus is a type for cluster status.
type ClusterStatus string

// HarborClusterConditionType is a valid value for HarborClusterConditionType.Type.
type HarborClusterConditionType string

// These are valid conditions of a HarborCluster.
const (
	// Ready means the HarborCluster is ready.
	Ready HarborClusterConditionType = "Ready"
	// CacheReady means the Cache is ready.
	CacheReady HarborClusterConditionType = "CacheReady"
	// DatabaseReady means the Database is ready.
	DatabaseReady HarborClusterConditionType = "DatabaseReady"
	// StorageReady means the Storage is ready.
	StorageReady HarborClusterConditionType = "StorageReady"
	// ServiceReady means the Service of Harbor is ready.
	ServiceReady HarborClusterConditionType = "ServiceReady"
	// ConfigurationReady means the configuration is applied to harbor.
	ConfigurationReady HarborClusterConditionType = "ConfigurationReady"
	// StatusCreating is the status of provisioning.
	StatusProvisioning ClusterStatus = "provisioning"
	// StatusHealthy is the status of healthy.
	StatusHealthy ClusterStatus = "healthy"
	// StatusUnHealthy is the status of unhealthy.
	StatusUnHealthy ClusterStatus = "unhealthy"
)

// HarborClusterCondition contains details for the current condition of this pod.
type HarborClusterCondition struct {
	// Type is the type of the condition.
	Type HarborClusterConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// the name of component used harbor cluster.
type Component string

// all Component used in harbor cluster full stack.
const (
	ComponentHarbor   Component = "harbor"
	ComponentCache    Component = "cache"
	ComponentStorage  Component = "storage"
	ComponentDatabase Component = "database"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=harborcluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.externalURL`,description="The public URL to the Harbor application",priority=0
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`,description="The overall status of the Harbor cluster",priority=0
// +kubebuilder:printcolumn:name="Operator Version",type=string,JSONPath=`.status.operator.controllerVersion`,description="The operator version ",priority=30
// +kubebuilder:printcolumn:name="Operator Git Commit",type=string,JSONPath=`.status.operator.controllerGitCommit`,description="The operator git commit",priority=30
// HarborCluster is the Schema for the harborclusters API.
type HarborCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborClusterSpec   `json:"spec,omitempty"`
	Status HarborClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// HarborClusterList contains a list of HarborCluster.
type HarborClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborCluster `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&HarborCluster{}, &HarborClusterList{})
}

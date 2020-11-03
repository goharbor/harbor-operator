package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	HarborClusterGVK = schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "HarborCluster",
	}
)

// HarborClusterSpec defines the desired state of HarborCluster
type HarborClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	HarborSpec `json:",inline"`

	// Cache configuration for in-cluster cache services
	// +optional
	InClusterCache *Cache `json:"inClusterCache,omitempty"`

	// Database configuration for in-cluster database service
	// +optional
	InClusterDatabase *Database `json:"inClusterDatabase,omitempty"`

	// Storage configuration for in-cluster storage service
	// +optional
	InClusterStorage *Storage `json:"inClusterStorage,omitempty"`

	// harbor version to be deployed, this version determines the image tags of harbor service components
	// +kubebuilder:validation:Required
	// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	// +kubebuilder:validation:Pattern="^(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
	Version string `json:"version"`
}

type Cache struct {
	// Set the kind of cache service to be used. Only support Redis now.
	// +kubebuilder:validation:Enum=Redis
	Kind string `json:"kind"`

	// +kubebuilder:validation:Required
	RedisSpec *RedisSpec `json:"redisSpec"`
}

type RedisSpec struct {
	Server   *RedisServer   `json:"server,omitempty"`
	Sentinel *RedisSentinel `json:"sentinel,omitempty"`

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int `json:"poolSize,omitempty"`

	// TLS Config to use. When set TLS will be negotiated.
	// set the secret which type of Opaque, and contains "tls.key","tls.crt","ca.crt".
	TlsConfig string `json:"tlsConfig,omitempty"`

	GroupName string `json:"groupName,omitempty"`

	// +kubebuilder:validation:Enum=sentinel;redis
	Schema string `json:"schema,omitempty"`

	Hosts []RedisHosts `json:"hosts,omitempty"`
}

type RedisHosts struct {
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
}

type RedisSentinel struct {
	Replicas int `json:"replicas,omitempty"`
}

type RedisServer struct {
	Replicas         int                         `json:"replicas,omitempty"`
	Resources        corev1.ResourceRequirements `json:"resources,omitempty"`
	StorageClassName string                      `json:"storageClassName,omitempty"`
	// the size of storage used in redis.
	Storage string `json:"storage,omitempty"`
}

type Database struct {
	// Set the kind of which database service to be used, Only support PostgresSQL now.
	// +kubebuilder:validation:Enum=PostgresSQL
	Kind string `json:"kind"`

	// +kubebuilder:validation:Required
	PostgresSQLSpec *PostgresSQLSpec `json:"postgresSqlSpec"`
}

type PostgresSQLSpec struct {
	Storage          string                      `json:"storage,omitempty"`
	Replicas         int                         `json:"replicas,omitempty"`
	Version          string                      `json:"version,omitempty"`
	StorageClassName string                      `json:"storageClassName,omitempty"`
	Resources        corev1.ResourceRequirements `json:"resources,omitempty"`
	SslConfig        string                      `json:"sslConfig,omitempty"`
	ConnectTimeout   int                         `json:"connectTimeout,omitempty"`
}

type Storage struct {
	// Kind of which storage service to be used. Only support MinIO now.
	// +kubebuilder:validation:Enum=MinIO
	Kind string `json:"kind"`

	// inCLuster options.
	MinIOSpec *MinIOSpec `json:"minIOSpec,omitempty"`
}

type MinIOSpec struct {
	// Supply number of replicas.
	// For standalone mode, supply 1. For distributed mode, supply 4 to 16 drives (should be even).
	// Note that the operator does not support upgrading from standalone to distributed mode.
	// +kubebuilder:validation:Required
	Replicas int32 `json:"replicas"`
	// Version defines the MinIO Client (mc) Docker image version.
	Version string `json:"version,omitempty"`
	// Number of persistent volumes that will be attached per server
	VolumesPerServer int32 `json:"volumesPerServer"`
	// VolumeClaimTemplate allows a user to specify how volumes inside a MinIOInstance
	// +optional
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// HarborClusterStatus defines the observed state of HarborCluster
type HarborClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []HarborClusterCondition `json:"conditions,omitempty"`
}

// HarborClusterConditionType is a valid value for HarborClusterConditionType.Type
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
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
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
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=0
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.externalURL`,description="The public URL to the Harbor application",priority=0
// +kubebuilder:printcolumn:name="Service Ready", type=string,JSONPath=`.status.conditions[?(@.type=="ServiceReady")].status`,description="The current status of the new Harbor spec",priority=10
// +kubebuilder:printcolumn:name="Cache Ready", type=string,JSONPath=`.status.conditions[?(@.type=="CacheReady")].status`,description="The current status of the new Cache spec",priority=20
// +kubebuilder:printcolumn:name="Database Ready", type=string,JSONPath=`.status.conditions[?(@.type=="DatabaseReady")].status`,description="The current status of the new Database spec",priority=20
// +kubebuilder:printcolumn:name="Storage Ready", type=string,JSONPath=`.status.conditions[?(@.type=="StorageReady")].status`,description="The current status of the new Storage spec",priority=20

// HarborCluster is the Schema for the harborclusters API
type HarborCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborClusterSpec   `json:"spec,omitempty"`
	Status HarborClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborClusterList contains a list of HarborCluster
type HarborClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborCluster `json:"items"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&HarborCluster{}, &HarborClusterList{})
}

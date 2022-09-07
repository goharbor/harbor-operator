package v1alpha3

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	HarborSpec `json:",inline"`

	// Cache configuration for in-cluster cache services
	// +kubebuilder:validation:Optional
	InClusterCache *Cache `json:"inClusterCache,omitempty"`

	// Database configuration for in-cluster database service
	// +kubebuilder:validation:Optional
	InClusterDatabase *Database `json:"inClusterDatabase,omitempty"`

	// Storage configuration for in-cluster storage service
	// +kubebuilder:validation:Optional
	InClusterStorage *Storage `json:"inClusterStorage,omitempty"`
}

type Cache struct {
	// Set the kind of cache service to be used. Only support Redis now.
	// +kubebuilder:validation:Enum=Redis
	Kind string `json:"kind"`

	// RedisSpec is the specification of redis.
	// +kubebuilder:validation:Required
	RedisSpec *RedisSpec `json:"redisSpec"`
}

type RedisSpec struct {
	harbormetav1.ImageSpec `json:",inline"`

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
	// Set the kind of which database service to be used, Only support PostgresSQL now.
	// +kubebuilder:validation:Enum=PostgresSQL
	Kind string `json:"kind"`

	// +kubebuilder:validation:Required
	PostgresSQLSpec *PostgresSQLSpec `json:"postgresSqlSpec"`
}

type PostgresSQLSpec struct {
	harbormetav1.ImageSpec `json:",inline"`

	// Storage defines database data store pvc size
	// +kubebuilder:validation:Optional
	Storage string `json:"storage,omitempty"`
	// Replicas defines database instance replicas
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum:=1
	Replicas int `json:"replicas,omitempty"`
	// StorageClassName defines use which StorageClass to create pvc
	// +kubebuilder:validation:Optional
	StorageClassName string `json:"storageClassName,omitempty"`
	// Resources defines database pod resource config
	// +kubebuilder:validation:Optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type Storage struct {
	// Kind of which storage service to be used. Only support MinIO now.
	// +kubebuilder:validation:Enum=MinIO
	Kind string `json:"kind"`

	// inCLuster options.
	// +kubebuilder:validation:Required
	MinIOSpec *MinIOSpec `json:"minIOSpec,omitempty"`
}

// StorageRedirectSpec defines if the redirection is disabled.
type StorageRedirectSpec struct {
	// Default is true
	// +kubebuilder:default:=true
	Enable bool `json:"enable"`
	// +kubebuilder:validation:Optional
	Expose *HarborExposeComponentSpec `json:"expose,omitempty"`
}

type MinIOSpec struct {
	harbormetav1.ImageSpec `json:",inline"`

	// Determine if the redirection of minio storage is disabled.
	// +kubebuilder:validation:Required
	Redirect StorageRedirectSpec `json:"redirect"`
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
// +k8s:openapi-gen=true
// +resource:path=harborcluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.externalURL`,description="The public URL to the Harbor application",priority=0
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The version to the Harbor application",priority=0
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

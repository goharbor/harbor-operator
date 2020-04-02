package v1alpha2

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CoreSecretKey       = "CORE_SECRET"
	CoreURLKey          = "CORE_URL"
	CoreRedisSessionKey = "_REDIS_URL"

	// CoreAdminUserKey     = corev1.BasicAuthUsernameKey
	CoreAdminPasswordKey = corev1.BasicAuthPasswordKey
)

const (
	CoreDatabaseHostKey     = "POSTGRESQL_HOST"
	CoreDatabasePortKey     = "POSTGRESQL_PORT"
	CoreDatabaseNameKey     = "POSTGRESQL_DATABASE"
	CoreDatabaseUserKey     = "POSTGRESQL_USERNAME"
	CoreDatabasePasswordKey = "POSTGRESQL_PASSWORD"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Core is the Schema for the registries API
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=core
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
type Core struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CoreSpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CoreList contains a list of Core
type CoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Core `json:"items"`
}

// CoreSpec defines the desired state of Core
type CoreSpec struct {
	ComponentSpec `json:",inline"`
	CoreConfig    `json:",inline"`

	// +kubebuilder:validation:Optional
	ReadOnly bool `json:"readOnly"`

	// +kubebuilder:validation:Required
	ConfigExpiration time.Duration `json:"configExpiration"`

	// +kubebuilder:validation:Optional
	RegistryCacheSecret string `json:"registryCacheSecret,omitempty"`

	// +kubebuilder:validation:Required
	AdminPasswordSecret string `json:"adminPasswordSecret"`

	// +kubebuilder:validation:Optional
	SessionRedisSecret string `json:"sessionRedisSecret,omitempty"`

	// +kubebuilder:validation:Optional
	ChartRepositoryURL string `json:"chartRepositoryURL"`

	// +kubebuilder:validation:Optional
	ClairAdapterURL string `json:"clairAdapterURL"`

	// +kubebuilder:validation:Optional
	ClairURL string `json:"clairURL"`

	// +kubebuilder:validation:Required
	ClairDatabaseSecret string `json:"clairDatabaseSecret"`

	// +kubebuilder:validation:Optional
	JobServiceURL string `json:"JobServiceURL"`

	// +kubebuilder:validation:Optional
	NotaryURL string `json:"notaryURL"`

	// +kubebuilder:validation:Optional
	RegistryURL string `json:"registryURL"`

	// +kubebuilder:validation:Optional
	RegistryControllerURL string `json:"registryControllerURL"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="INFO,DEBUG,WARNING,ERROR,FATAL"
	LogLevel string `json:"logLevel,omitempty"`

	// +kubebuilder:validation:Optional
	SyncRegistry bool `json:"syncRegistry"`

	// +kubebuilder:validation:Optional
	SyncQuota bool `json:"syncQuota"`

	// +kubebuilder:validation:Required
	PublicURL string `json:"publicURL"`
}

type CoreConfig struct {
	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`
}

// nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&Core{}, &CoreList{})
}

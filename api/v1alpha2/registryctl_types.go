package v1alpha2

import (
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ipaddress:port[,weight,password,database_index]
	RegistryControllerCacheURLKey = "url"
)

const (
	RegistryControllerCorePublicURLKey = "REGISTRY_HTTP_HOST"
	RegistryControllerAuthURLKey       = "REGISTRY_AUTH_TOKEN_REALM" // RegistryControllerCorePublicURLKey + "/service/token"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RegistryController is the Schema for the registries API
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=registrycontroller
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
type RegistryController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RegistryControllerSpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegistryControllerList contains a list of RegistryController
type RegistryControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryController `json:"items"`
}

// RegistryControllerSpec defines the desired state of RegistryController
type RegistryControllerSpec struct {
	ComponentSpec               `json:",inline"`
	RegistryControllerComponent `json:",inline"`

	StorageSecret string `json:"storageSecret,omitempty"`

	CacheSecret string `json:"cacheSecret,omitempty"`

	CoreSecret string `json:"coreSecret,omitempty"`

	JobServiceSecret string `json:"jobService,omitempty"`

	ConfigName string `json:"configName"`

	PublicURL string `json:"publicURL"`

	// The issuer for Harbor certificates.
	// If the 'kind' field is not set, or set to 'Issuer', an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	CertificateIssuerRef cmmeta.ObjectReference `json:"certificateIssuerRef"`
}

// nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&RegistryController{}, &RegistryControllerList{})
}

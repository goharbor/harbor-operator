package v1alpha2

import (
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ipaddress:port[,weight,password,database_index]
	RegistryCacheURLKey = "url"
)

const (
	RegistryCorePublicURLKey = "REGISTRY_HTTP_HOST"
	RegistryAuthURLKey       = "REGISTRY_AUTH_TOKEN_REALM" // RegistryCorePublicURLKey + "/service/token"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Registry is the Schema for the registries API
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=registry
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RegistrySpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegistryList contains a list of Registry
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

// RegistrySpec defines the desired state of Registry
type RegistrySpec struct {
	ComponentSpec     `json:",inline"`
	RegistryComponent `json:",inline"`

	// +kubebuilder:validation:Optional
	StorageSecret string `json:"storageSecret,omitempty"`

	CacheSecret string `json:"cacheSecret,omitempty"`

	CoreSecret string `json:"coreSecret,omitempty"`

	JobServiceSecret string `json:"jobService,omitempty"`

	// +kubebuilder:validation:Required
	ConfigName string `json:"configName"`

	// +kubebuilder:validation:Required
	PublicURL string `json:"publicURL"`

	// The issuer for Harbor certificates.
	// If the 'kind' field is not set, or set to 'Issuer', an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	// +kubebuilder:validation:Required
	CertificateIssuerRef cmmeta.ObjectReference `json:"certificateIssuerRef"`
}

// nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}

package v1alpha2

import (
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

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=registrycontroller
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// RegistryController is the Schema for the registries API.
type RegistryController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RegistryControllerSpec `json:"spec,omitempty"`

	// Most recently observed status.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// RegistryControllerList contains a list of RegistryController.
type RegistryControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryController `json:"items"`
}

// RegistryControllerSpec defines the desired state of RegistryController.
type RegistryControllerSpec struct {
	ComponentSpec               `json:",inline"`
	RegistryControllerComponent `json:",inline"`

	// +kubebuilder:validation:Required
	RegistryRef string `json:"registryRef"`

	// +kubebuilder:validation:Optional
	Log RegistryControllerLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	HTTPS RegistryControllerHTTPSSpec `json:"https,omitempty"`
}

type RegistryControllerLogSpec struct {
	// +kubebuilder:validation:Optional
	Level string `json:"level,omitempty"`
}

type RegistryControllerHTTPSSpec struct {
	// +kubebuilder:validation:Optional
	CertificateRef string `json:"certificateRef,omitempty"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&RegistryController{}, &RegistryControllerList{})
}

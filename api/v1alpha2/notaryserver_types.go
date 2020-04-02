package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NotaryServer is the Schema for the registries API
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=notaryserver
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
type NotaryServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotaryServerSpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotaryServerList contains a list of NotaryServer
type NotaryServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotaryServer `json:"items"`
}

// NotaryServerSpec defines the desired state of NotaryServer
type NotaryServerSpec struct {
	ComponentSpec         `json:",inline"`
	NotaryServerComponent `json:",inline"`

	// The url exposed to clients to access notary
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	PublicURL string `json:"publicURL"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	SignerURL string `json:"signerURL"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`

	// +kubebuilder:validation:Required
	CertificateSecret string `json:"certificateSecret"`

	// +kubebuilder:validation:Required
	TokenSecret string `json:"tokenSecret"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&NotaryServer{}, &NotaryServerList{})
}

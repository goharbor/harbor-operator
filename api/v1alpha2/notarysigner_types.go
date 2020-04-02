package v1alpha2

import (
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NotarySigner is the Schema for the registries API
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=notarysigner
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
type NotarySigner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotarySignerSpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotarySignerList contains a list of NotarySigner
type NotarySignerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotarySigner `json:"items"`
}

// NotarySignerSpec defines the desired state of NotarySigner
type NotarySignerSpec struct {
	ComponentSpec         `json:",inline"`
	NotarySignerComponent `json:",inline"`

	// The url exposed to clients to access notary
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	PublicURL string `json:"publicURL"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`

	// +kubebuilder:validation:Required
	CertificateSecret string `json:"certificateSecret"`

	// The issuer for Harbor certificates.
	// If the 'kind' field is not set, or set to 'Issuer', an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	CertificateIssuerRef cmmeta.ObjectReference `json:"certificateIssuerRef"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&NotarySigner{}, &NotarySignerList{})
}

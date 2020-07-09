package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=notarysigner
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// NotarySigner is the Schema for the registries API.
type NotarySigner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotarySignerSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// NotarySignerList contains a list of NotarySigner.
type NotarySignerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotarySigner `json:"items"`
}

// NotarySignerSpec defines the desired state of NotarySigner.
type NotarySignerSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	// The url exposed to clients to access notary
	PublicURL string `json:"publicURL"`

	// +kubebuilder:validation:Required
	HTTPS NotaryHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Optional
	Logging NotaryLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Required
	Storage NotarySignerStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Migration *NotaryMigrationSpec `json:"migration,omitempty"`
}

type NotarySignerStorageSpec struct {
	NotaryStorageSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	AliasesRef string `json:"aliasesRef"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&NotarySigner{}, &NotarySignerList{})
}

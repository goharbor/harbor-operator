package v1alpha2

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NotarySignerAPIPort = 7899

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=notarysigner
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// NotarySigner is the Schema for the NotarySigner API.
type NotarySigner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotarySignerSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
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
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Required
	Authentication NotarySignerAuthenticationSpec `json:"authentatication"`

	// +kubebuilder:validation:Optional
	Logging NotaryLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Required
	Storage NotarySignerStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Migration *NotaryMigrationSpec `json:"migration,omitempty"`
}

type NotarySignerAuthenticationSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef"`
}

type NotarySignerStorageSpec struct {
	NotaryStorageSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	AliasesRef string `json:"aliasesRef"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&NotarySigner{}, &NotarySignerList{})
}

package v1alpha2

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=clair
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// Clair is the Schema for the Clair API.
type Clair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClairSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClairList contains a list of Clair.
type ClairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Clair `json:"items"`
}

// ClairSpec defines the desired state of Clair.
type ClairSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +listType:set
	VulnerabilitySources []string `json:"vulnerabilitySources"`

	// +kubebuilder:validation:Required
	Adapter ClairAdapterComponent `json:"adapter"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`
}

type ClairAdapterComponent struct {
	// +kubebuilder:validation:Required
	Redis OpacifiedDSN `json:"redis"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Clair{}, &ClairList{})
}

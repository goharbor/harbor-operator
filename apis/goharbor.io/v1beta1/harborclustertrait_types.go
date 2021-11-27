package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HarborClusterTraitSpec defines the desired state of HarborClusterTrait.
type HarborClusterTraitSpec struct {
	// +kubebuilder:validation:Optional
	Affinities []Affinity `json:"affinities,omitempty"`
}

type Affinity struct {
	// +kubebuilder:validation:Optional
	// LabelSelector is use label to fields pod
	Selector LabelSelector `json:"selector"`

	// +kubebuilder:validation:Optional
	// Affinity is the configuration of the pod affinity.
	Affinity *corev1.Affinity `json:"affinity"`
}

type LabelSelector struct {
	// matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
	// map is equivalent to an element of matchExpressions, whose key field is "key", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// AffinityStore is to store label and affinity mapping.
type AffinityStore struct {
	SelectLabels map[string]*corev1.Affinity
}

// HarborClusterTraitStatus defines the observed state of HarborClusterTrait.
type HarborClusterTraitStatus struct {
	// Status represents harbor configuration status.
	// +kubebuilder:validation:Optional
	Status HarborServerConfigurationStatusType `json:"status,omitempty"`
	// Reason represents status reason.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`
	// Message provides human-readable message.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true

// HarborClusterTrait is the Schema for the harborclustertraits API.
type HarborClusterTrait struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborClusterTraitSpec   `json:"spec,omitempty"`
	Status HarborClusterTraitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborClusterTraitList contains a list of HarborClusterTrait.
type HarborClusterTraitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborClusterTrait `json:"items"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&HarborClusterTrait{}, &HarborClusterTraitList{})
}

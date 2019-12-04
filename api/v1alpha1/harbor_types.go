/*
.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HarborSpec defines the desired state of Harbor
type HarborSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Harbor. Edit Harbor_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// HarborStatus defines the observed state of Harbor
type HarborStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Harbor is the Schema for the harbors API
type Harbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborSpec   `json:"spec,omitempty"`
	Status HarborStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborList contains a list of Harbor
type HarborList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Harbor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}

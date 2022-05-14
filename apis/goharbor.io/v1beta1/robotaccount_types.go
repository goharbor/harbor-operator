package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RobotAccountSpec defines the desired state of RobotAccount
type RobotAccountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of RobotAccount. Edit robotaccount_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// RobotAccountStatus defines the observed state of RobotAccount
type RobotAccountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RobotAccount is the Schema for the robotaccounts API
type RobotAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RobotAccountSpec   `json:"spec,omitempty"`
	Status RobotAccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RobotAccountList contains a list of RobotAccount
type RobotAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RobotAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RobotAccount{}, &RobotAccountList{})
}

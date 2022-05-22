package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type RobotBodyPermission struct {
	// +kubebuilder:validation:Required
	Access []RobotBodyAccess `json:"access,omitempty"`
	// +kubebuilder:validation:Required
	Kind string `json:"kind,omitempty"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace,omitempty"`
}
type RobotBodyAccess struct {
	// +kubebuilder:validation:Required
	Action string `json:"action,omitempty"`
	// +kubebuilder:validation:Required
	Resource string `json:"resource,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// RobotAccountSpec defines the desired state of RobotAccount.
type RobotAccountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	Level string `json:"level"`

	// Indicate which harbor server configuration is referred
	// +kubebuilder:validation:Required
	HarborServerConfig string `json:"harborServerConfig"`

	// +kubebuilder:validation:Required
	Permissions []RobotBodyPermission `json:"permissions,omitempty"`

	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Optional
	Duration int `json:"duration,omitempty"`

	// +kubebuilder:validation:Optional
	Disable bool `json:"disable,omitempty"`
}

// RobotAccountStatus defines the observed state of RobotAccount.
type RobotAccountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Reason represents status reason.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`
	// Message provides human-readable message.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
	// ID is robot account id.
	// +kubebuilder:validation:Optional
	ID int64 `json:"id,omitempty"`
	// Secret is robot account secret.
	// +kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RobotAccount is the Schema for the robotaccounts API.
type RobotAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status RobotAccountStatus `json:"status,omitempty"`
	Spec   RobotAccountSpec   `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// RobotAccountList contains a list of RobotAccount.
type RobotAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RobotAccount `json:"items"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&RobotAccount{}, &RobotAccountList{})
}

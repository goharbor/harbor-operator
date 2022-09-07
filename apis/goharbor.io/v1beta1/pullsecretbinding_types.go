package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PullSecretBindingSpec defines the desired state of PullSecretBinding.
type PullSecretBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// RobotID points to the robot account id used for secret binding
	// +kubebuilder:validation:Required
	RobotID string `json:"robotId"`

	// ProjectID points to the project associated with the secret binding
	// +kubebuilder:validation:Required
	ProjectID string `json:"projectId"`

	// Indicate which harbor server configuration is referred
	HarborServerConfig string `json:"harborServerConfig"`

	// Indicate which service account binds the pull secret
	ServiceAccount string `json:"serviceAccount"`
}

// PullSecretBindingStatusType defines the status type of configuration.
type PullSecretBindingStatusType string

const (
	// PullSecretBindingStatusBinding represents ready status.
	PullSecretBindingStatusBinding PullSecretBindingStatusType = "Binding"
	// PullSecretBindingStatusBound represents fail status.
	PullSecretBindingStatusBound PullSecretBindingStatusType = "Bound"
	// PullSecretBindingStatusUnknown represents unknown status.
	PullSecretBindingStatusUnknown PullSecretBindingStatusType = "Unknown"
)

// PullSecretBindingStatus defines the observed state of PullSecretBinding.
type PullSecretBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Indicate the status of binding: `binding`, `bound` and `unknown`
	Status PullSecretBindingStatusType `json:"status"`
	// Reason represents status reason.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`
	// Message provides human-readable message.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="psb"
// +kubebuilder:printcolumn:name="Harbor Server",type=string,JSONPath=`.spec.harborServerConfig`,description="The Harbor server configuration CR reference",priority=0
// +kubebuilder:printcolumn:name="Service Account",type=string,JSONPath=`.spec.serviceAccount`,description="The service account binding the pull secret",priority=0
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`,description="The status of the Harbor server",priority=0

// PullSecretBinding is the Schema for the pullsecretbindings API.
type PullSecretBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PullSecretBindingSpec   `json:"spec,omitempty"`
	Status PullSecretBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PullSecretBindingList contains a list of PullSecretBinding.
type PullSecretBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PullSecretBinding `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&PullSecretBinding{}, &PullSecretBindingList{})
}

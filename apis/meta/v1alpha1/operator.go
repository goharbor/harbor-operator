package v1alpha1

// +kubebuilder:validation:Type=object
// ControllerStatus represents the current status of the operator.
type OperatorStatus struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	ControllerGitCommit string `json:"controllerGitCommit,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	ControllerVersion string `json:"controllerVersion,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ControllerName string `json:"controllerName,omitempty"`
}

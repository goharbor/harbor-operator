package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HarborServerConfigurationSpec defines the desired state of HarborServerConfiguration.
type HarborServerConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^https?://(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)+([A-Za-z]|[A-Za-z][A-Za-z0-9\\-]*[A-Za-z0-9])"
	ServerURL string `json:"serverURL"`

	// Indicate if the Harbor server is an insecure registry
	// +kubebuilder:validation:Optional
	Insecure bool `json:"insecure,omitempty"`

	// Default indicates the harbor configuration manages namespaces.
	// Value in goharbor.io/harbor annotation will be considered with high priority.
	// At most, one HarborServerConfiguration can be the default, multiple defaults will be rejected.
	// +kubebuilder:validation:Required
	Default bool `json:"default,omitempty"`

	// +kubebuilder:validation:Required
	AccessCredential *AccessCredential `json:"accessCredential"`

	// The version of the Harbor server
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"
	Version string `json:"version"`

	// Rules configures the container image rewrite rules for transparent proxy caching with Harbor.
	// +kubebuilder:validation:Optional
	Rules []string `json:"rules,omitempty"`

	// NamespaceSelector decides whether to apply the HSC on a namespace based
	// on whether the namespace matches the selector.
	// See
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
	// for more examples of label selectors.
	//
	// Default to the empty LabelSelector, which matches everything.
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

// AccessCredential is a namespaced credential to keep the access key and secret for the harbor server configuration.
type AccessCredential struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	AccessSecretRef string `json:"accessSecretRef"`
}

// HarborServerConfigurationStatusType defines the status type of configuration.
type HarborServerConfigurationStatusType string

const (
	// HarborServerConfigurationStatusReady represents ready status.
	HarborServerConfigurationStatusReady HarborServerConfigurationStatusType = "Success"
	// HarborServerConfigurationStatusFail represents fail status.
	HarborServerConfigurationStatusFail HarborServerConfigurationStatusType = "Fail"
	// HarborServerConfigurationStatusUnknown represents unknown status.
	HarborServerConfigurationStatusUnknown HarborServerConfigurationStatusType = "Unknown"
)

// HarborConfigurationStatus defines the status of HarborServerConfiguration.
type HarborServerConfigurationStatus struct {
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
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="hsc",scope="Cluster"
// +kubebuilder:printcolumn:name="Harbor Server",type=string,JSONPath=`.spec.serverURL`,description="The public URL to the Harbor server",priority=0
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`,description="The status of the Harbor server",priority=0
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The version of the Harbor server",priority=5
// HarborServerConfiguration is the Schema for the harborserverconfigurations API.
type HarborServerConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborServerConfigurationSpec   `json:"spec,omitempty"`
	Status HarborServerConfigurationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborServerConfigurationList contains a list of HarborServerConfiguration.
type HarborServerConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborServerConfiguration `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&HarborServerConfiguration{}, &HarborServerConfigurationList{})
}

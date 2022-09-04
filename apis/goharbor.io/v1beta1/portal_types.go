package v1beta1

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=portal
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// Portal is the Schema for the portals API.
type Portal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PortalSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// PortalList contains a list of Portal.
type PortalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Portal `json:"items"`
}

// PortalSpec defines the desired state of Portal.
type PortalSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1024
	// Sets the maximum number of simultaneous connections that can be opened.
	MaxConnections *int32 `json:"maxConnections,omitempty"`

	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	// Network settings for the portal
	Network *harbormetav1.Network `json:"network,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Portal{}, &PortalList{})
}

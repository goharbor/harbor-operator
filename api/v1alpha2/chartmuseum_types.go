package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ChartMuseumStorageKindKey = "kind"
)

const (
	ChartMuseumCacheURLKey = "url"
)

const (
	ChartMuseumBasicAuthKey = "BASIC_AUTH_PASS"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChartMuseum is the Schema for the registries API
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=chartmuseum
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
type ChartMuseum struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChartMuseumSpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChartMuseumList contains a list of ChartMuseum
type ChartMuseumList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChartMuseum `json:"items"`
}

// ChartMuseumSpec defines the desired state of ChartMuseum
type ChartMuseumSpec struct {
	ComponentSpec        `json:",inline"`
	ChartMuseumComponent `json:",inline"`

	StorageSecret string `json:"storageSecret,omitempty"`

	CacheSecret string `json:"cacheSecret,omitempty"`

	SecretName string `json:"secret,omitempty"`

	// The url exposed to clients to access ChartMuseum (probably https://.../chartrepo)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	PublicURL string `json:"publicURL"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&ChartMuseum{}, &ChartMuseumList{})
}

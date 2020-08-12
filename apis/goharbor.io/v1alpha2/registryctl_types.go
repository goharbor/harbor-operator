package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

const (
	// ipaddress:port[,weight,password,database_index]
	RegistryControllerCacheURLKey = "url"
)

const (
	RegistryControllerCorePublicURLKey = "REGISTRY_HTTP_HOST"
	RegistryControllerAuthURLKey       = "REGISTRY_AUTH_TOKEN_REALM" // RegistryControllerCorePublicURLKey + "/service/token"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=registrycontroller
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// RegistryController is the Schema for the registries API.
type RegistryController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RegistryControllerSpec `json:"spec,omitempty"`

	// Most recently observed status.
	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// RegistryControllerList contains a list of RegistryController.
type RegistryControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryController `json:"items"`
}

// RegistryControllerSpec defines the desired state of RegistryController.
type RegistryControllerSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	RegistryRef string `json:"registryRef"`

	// +kubebuilder:validation:Optional
	Log RegistryControllerLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Required
	Authentication RegistryControllerAuthenticationSpec `json:"authentication"`
}

type RegistryControllerAuthenticationSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CoreSecretRef string `json:"coreSecretRef,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	JobServiceSecretRef string `json:"jobServiceSecretRef,omitempty"`
}

type RegistryControllerLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	Level harbormetav1.RegistryCtlLogLevel `json:"level,omitempty"`
}

type RegistryControllerHTTPSSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&RegistryController{}, &RegistryControllerList{})
}

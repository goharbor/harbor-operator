package v1alpha3

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// +k8s:openapi-gen=true
// +resource:path=registrycontroller
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// RegistryController is the Schema for the RegistryController API.
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

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&RegistryController{}, &RegistryControllerList{})
}

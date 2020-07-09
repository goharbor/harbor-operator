package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=notaryserver
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// NotaryServer is the Schema for the registries API.
type NotaryServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotaryServerSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// NotaryServerList contains a list of NotaryServer.
type NotaryServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotaryServer `json:"items"`
}

// NotaryServerSpec defines the desired state of NotaryServer.
type NotaryServerSpec struct {
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	HTTPS *NotaryHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Required
	TrustService NotaryServerTrustServiceSpec `json:"trustService"`

	// +kubebuilder:validation:Optional
	Logging NotaryLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Required
	Storage NotaryStorageSpec `json:"storage,omitempty"`

	// +kubebuilder:validation:Optional
	Auth NotaryServerAuthSpec `json:"auth,omitempty"`

	// +kubebuilder:validation:Optional
	Migration *NotaryMigrationSpec `json:"migration,omitempty"`
}

type NotaryServerTrustServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"remote","local"}
	Type string `json:"type"`

	// +kubebuilder:validation:Optional
	Host string `json:"host,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:default=443
	Port int64 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"ecdsa","rsa","ed25519"}
	// +kubebuilder:default=ecdsa
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef,omitempty"`
}

type NotaryServerAuthSpec struct {
	// +kubebuilder:validation:Required
	Token NotaryServerAuthTokenSpec `json:"token"`
}

type NotaryServerAuthTokenSpec struct {
	// +kubebuilder:validation:Required
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	Service string `json:"service"`

	// +kubebuilder:validation:Required
	Issuer string `json:"issuer"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	AutoRedirect *bool `json:"autoredirect,omitempty"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&NotaryServer{}, &NotaryServerList{})
}

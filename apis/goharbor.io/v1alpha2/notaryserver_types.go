package v1alpha2

import (
	"errors"

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
	HTTPS NotaryServerHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Required
	TrustService NotaryServerTrustServiceSpec `json:"trustService"`

	// +kubebuilder:validation:Required
	Storage NotaryServerStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Logging NotaryLoggingSpec `json:"logging"`

	// +kubebuilder:validation:Optional
	Auth NotaryServerAuthSpec `json:"auth"`

	// +kubebuilder:validation:Optional
	Migration NotaryMigrationSpec `json:"migration"`
}

type NotaryServerHTTPSSpec struct {
	// +kubebuilder:validation:Required
	CertificateRef string `json:"certificateRef"`
}

type NotaryServerTrustServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"remote","local"}
	Type string `json:"type"`

	// +kubebuilder:validation:Optional
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	Port int64 `json:"port"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"ecdsa","rsa","ed25519"}
	// +kubebuilder:default=ecdsa
	KeyAlgorithm string `json:"keyAlgorithm"`

	// +kubebuilder:validation:Optional
	CertificateRef string `json:"certificateRef"`
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
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	AutoRedirect bool `json:"autoredirect"`
}

type NotaryServerStorageSpec struct {
	OpacifiedDSN `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"mysql","postgres","memory"}
	Type string `json:"type"`
}

var (
	errNotImplemented = errors.New("not yet implemented")
)

func (n *NotaryServerStorageSpec) GetPasswordFieldKey() (string, error) {
	switch n.Type {
	case "postgres":
		return PostgresqlPasswordKey, nil
	default:
		return "", errNotImplemented
	}
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&NotaryServer{}, &NotaryServerList{})
}

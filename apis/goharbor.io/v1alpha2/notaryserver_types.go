package v1alpha2

import (
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=notaryserver
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// NotaryServer is the Schema for the NotaryServer API.
type NotaryServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotaryServerSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
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
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Required
	TrustService NotaryServerTrustServiceSpec `json:"trustService"`

	// +kubebuilder:validation:Optional
	Logging NotaryLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Required
	Storage NotaryStorageSpec `json:"storage,omitempty"`

	// +kubebuilder:validation:Optional
	Authentication *NotaryServerAuthSpec `json:"authentication,omitempty"`

	// +kubebuilder:validation:Optional
	Migration *NotaryMigrationSpec `json:"migration,omitempty"`
}

const (
	NotaryServerTrustRemoteType = "remote"
	NotaryServerTrustLocalType  = "local"
)

type NotaryServerTrustServiceSpec struct {
	// +kubebuilder:validation:Optional
	Remote *NotaryServerTrustServiceRemoteSpec `json:"remote,omitempty"`
}

type NotaryServerTrustServiceRemoteSpec struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:default=443
	Port int64 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ecdsa
	KeyAlgorithm certv1.KeyAlgorithm `json:"keyAlgorithm,omitempty"`

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

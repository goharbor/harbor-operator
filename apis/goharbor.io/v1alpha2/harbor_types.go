package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=harbor
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="h"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.publicURL`,description="The public URL to the Harbor application",priority=0
// Harbor is the Schema for the harbors API.
type Harbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HarborSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +resource:path=harbors
// HarborList contains a list of Harbor.
type HarborList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Harbor `json:"items"`
}

// HarborSpec defines the desired state of Harbor.
type HarborSpec struct {
	// The Harbor semver version
	// +kubebuilder:validation:Required
	// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	// +kubebuilder:validation:Pattern="^(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
	HarborVersion string `json:"version"`

	// The url exposed to clients to access harbor
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	PublicURL string `json:"publicURL"`

	// The name of the secret containing the TLS secret used for ingresses
	TLSSecretName string `json:"tlsSecretName"`

	// +kubebuilder:validation:Required
	Components HarborComponents `json:"components,omitempty"`

	// The name of the secret containing the password for root user
	// +kubebuilder:validation:Required
	AdminPasswordSecret string `json:"adminPasswordSecret"`

	// The Maximum priority. Deployments may be created with priority in interval ] priority - 100 ; priority ]
	// +kubebuilder:validation:Optional
	Priority *int32 `json:"priority,omitempty"`

	// The option to set repository read only.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	ReadOnly bool `json:"readOnly,omitempty"`

	// Indicates that the harbor is paused.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Paused bool `json:"paused,omitempty"`

	// The issuer for Harbor certificates.
	// If the 'kind' field is not set, or set to 'Issuer', an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	CertificateIssuerRef cmmeta.ObjectReference `json:"certificateIssuerRef"`
}

type CoreComponent struct {
	CoreConfig `json:",inline"`
}
type PortalComponent struct{}
type RegistryComponent struct{}
type RegistryControllerComponent struct{}
type JobServiceComponent struct{}
type ChartMuseumComponent struct{}
type NotaryServerComponent struct {
	// The url exposed to clients to access notary
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	PublicURL string `json:"publicURL"`
}
type NotarySignerComponent struct {
	// CommonName is a common name to be used on the Certificate.
	// The CommonName should have a length of 64 characters or fewer to avoid
	// generating invalid CSRs.
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Optional
	CommonName string `json:"commonName,omitempty"`

	// Organization is the organization to be used on the Certificate
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// This cannot be set to true: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#validation
	// +listType:atomic
	Organization []string `json:"organization"`

	// KeySize is the key bit size of the corresponding private key for this certificate.
	// +optional
	// +kubebuilder:validation:Maximum=8192
	// +kubebuilder:validation:Minimum=2048
	KeySize int `json:"keySize,omitempty"`
}

type ClairAdapterComponent struct {
	// +kubebuilder:validation:Required
	RedisSecret string `json:"redisSecret"`
}

type ClairComponent struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +listType:set
	VulnerabilitySources []string `json:"vulnerabilitySources"`

	// +kubebuilder:validation:Required
	Adapter ClairAdapterComponent `json:"adapter"`
}
type HarborComponents struct {
	Core *CoreComponent `json:"core"`

	Portal *PortalComponent `json:"portal"`

	Registry *RegistryComponent `json:"registry"`

	RegistryController *RegistryControllerComponent `json:"registryController"`

	JobService *JobServiceComponent `json:"jobService"`

	ChartMuseum *ChartMuseumComponent `json:"chartMuseum"`

	Clair *ClairComponent `json:"clair"`

	NotaryServer *NotaryServerComponent `json:"notaryServer"`
	NotarySigner *NotarySignerComponent `json:"notarySigner"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}

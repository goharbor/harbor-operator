package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Harbor is the Schema for the harbors API
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +resource:path=harbor
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="h"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Public URL",type=string,JSONPath=`.spec.publicURL`,description="The public URL to the Harbor application",priority=0
// +kubebuilder:printcolumn:name="Applied",type=string,JSONPath=`.status.conditions[?(@.type=="Applied")].status`,description="The current status of the new Harbor spec",priority=20
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="The current status of the Harbor application",priority=10
type Harbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HarborSpec `json:"spec,omitempty"`

	// Most recently observed status of the Harbor.
	// +optional
	Status HarborStatus `json:"status,omitempty"`
}

// HarborList contains a list of Harbor
// +kubebuilder:object:root=true
// +resource:path=harbors
type HarborList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Harbor `json:"items"`
}

// HarborSpec defines the desired state of Harbor
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
	// +optional
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
	ReadOnly bool `json:"readOnly,omitempty"`

	// Indicates that the harbor is paused.
	// +optional
	Paused bool `json:"paused,omitempty"`

	// The issuer for Harbor certificates.
	// If the 'kind' field is not set, or set to 'Issuer', an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	CertificateIssuerRef cmmeta.ObjectReference `json:"certificateIssuerRef"`
}

type HarborComponents struct {
	// +optional
	Core *CoreComponent `json:"core,omitempty"`

	// +optional
	Portal *PortalComponent `json:"portal,omitempty"`

	// +optional
	Registry *RegistryComponent `json:"registry,omitempty"`

	// +optional
	JobService *JobServiceComponent `json:"jobService,omitempty"`

	// +optional
	ChartMuseum *ChartMuseumComponent `json:"chartMuseum,omitempty"`

	// +optional
	Clair *ClairComponent `json:"clair,omitempty"`

	// +optional
	Notary *NotaryComponent `json:"notary,omitempty"`
}

type HarborDeployment struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Image *string `json:"image,omitempty"`

	// +optional
	NodeSelector     NodeSelector                  `json:"nodeSelector,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

type NodeSelector map[string]string

type CoreComponent struct {
	HarborDeployment `json:",inline"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`
}

type PortalComponent struct {
	HarborDeployment `json:",inline"`
}

type RegistryComponent struct {
	HarborDeployment `json:",inline"`

	Controller RegistryControllerComponent `json:"controller,omitempty"`

	// +optional
	StorageSecret string `json:"storageSecret,omitempty"`

	// +optional
	CacheSecret string `json:"cacheSecret,omitempty"`
}

type RegistryControllerComponent struct {
	// +optional
	Image *string `json:"image,omitempty"`
}

type JobServiceComponent struct {
	HarborDeployment `json:",inline"`

	// +kubebuilder:validation:Required
	RedisSecret string `json:"redisSecret"`

	// +optional
	WorkerCount int32 `json:"workerCount"`
}

type ClairAdapterComponent struct {
	// +optional
	Image *string `json:"image,omitempty"`

	// +kubebuilder:validation:Required
	RedisSecret string `json:"redisSecret"`
}

type ClairComponent struct {
	HarborDeployment `json:",inline"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`

	VulnerabilitySources []string `json:"vulnerabilitySources"`

	Adapter ClairAdapterComponent `json:"adapter"`
}

type ChartMuseumComponent struct {
	HarborDeployment `json:",inline"`

	// +optional
	StorageSecret string `json:"storageSecret,omitempty"`

	// +optional
	CacheSecret string `json:"cacheSecret,omitempty"`
}

type NotaryComponent struct {
	// The url exposed to clients to access notary
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	PublicURL string `json:"publicURL"`

	// +optional
	DBMigrator NotaryDBMigrator `json:"dbMigrator,omitempty"`

	// +kubebuilder:validation:Required
	Signer NotarySignerComponent `json:"signer"`

	// +kubebuilder:validation:Required
	Server NotaryServerComponent `json:"server"`
}

type NotaryDBMigrator struct {
	// +optional
	Image *string `json:"image,omitempty"`
}

type NotarySignerComponent struct {
	HarborDeployment `json:",inline"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`
}

type NotaryServerComponent struct {
	HarborDeployment `json:",inline"`

	// +kubebuilder:validation:Required
	DatabaseSecret string `json:"databaseSecret"`
}

// HarborStatus defines the observed state of Harbor
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
type HarborStatus struct {
	// Represents the latest available observations of a harbor's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []HarborCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,6,rep,name=conditions"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// HarborCondition describes the state of a Harbor at a certain point.
type HarborCondition struct {
	// Type of harhor condition.
	Type HarborConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type HarborConditionType string

const (
	AppliedConditionType HarborConditionType = "Applied"
	ReadyConditionType   HarborConditionType = "Ready"
)

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}

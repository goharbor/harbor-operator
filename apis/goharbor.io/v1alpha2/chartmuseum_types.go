package v1alpha2

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=chartmuseum
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// ChartMuseum is the Schema for the registries API.
type ChartMuseum struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChartMuseumSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ChartMuseumList contains a list of ChartMuseum.
type ChartMuseumList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChartMuseum `json:"items"`
}

// ChartMuseumSpec defines the desired state of ChartMuseum.
type ChartMuseumSpec struct {
	ComponentSpec        `json:",inline"`
	ChartMuseumComponent `json:",inline"`

	// +kubebuilder:validation:Optional
	Log ChartMuseumLogSpec `json:"log"`

	// +kubebuilder:validation:Optional
	Auth ChartMuseumAuthSpec `json:"auth"`

	// +kubebuilder:validation:Optional
	Server ChartMuseumServerSpec `json:"server"`

	// +kubebuilder:validation:Optional
	// Disable some features
	Disable ChartMuseumDisableSpec `json:"disable"`

	// +kubebuilder:validation:Optional
	// Cache stores
	Cache ChartMuseumCacheSpec `json:"cache"`

	// +kubebuilder:validation:Required
	Chart ChartMuseumChartSpec `json:"chart"`
}

type ChartMuseumServerSpec struct {
	// +kubebuilder:validation:Optional
	HTTPS ChartMuseumHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Optional
	// Socket timeout in nanoseconds
	ReadTimeout time.Duration `json:"readTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// Socket timeout in nanoseconds
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=20971520
	// Max size of post body (in bytes)
	MaxUploadSize int64 `json:"maxUploadSize"`

	// +kubebuilder:validation:Optional
	// Value to set in the Access-Control-Allow-Origin HTTP header
	CORSAllowOrigin string `json:"corsAllowOrigin,omitempty"`
}

type ChartMuseumChartSpec struct {
	// +kubebuilder:validation:Optional
	// Form fields which will be queried
	PostFormFieldName ChartMuseumPostFormFieldNameSpec `json:"postFormFieldName,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*$"
	// The absolute url for .tgzs in index.yaml
	URL string `json:"url"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// Allow chart versions to be re-uploaded without ?force querystring
	AllowOvewrite bool `json:"allowOverwrite"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Enforce the chart museum server only accepts the valid chart version as Helm does
	SemanticVersioning2Only bool `json:"onlySemver2"`

	// +kubebuilder:validation:Required
	Storage ChartMuseumChartStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Index ChartMuseumChartIndexSpec `json:"index"`

	// +kubebuilder:validation:Optional
	Repo ChartMuseumChartRepoSpec `json:"repo"`
}

type ChartMuseumChartRepoSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The length of repo variable
	DepthDynamic bool `json:"depthDynamic"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// Levels of nested repos for multitenancy
	Depth int `json:"depth"`
}

type ChartMuseumChartStorageSpec struct {
	ChartMuseumChartStorageDriverSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// Maximum number of objects allowed in storage (per tenant)
	MaxStorageObjects int64 `json:"maxStorageObject,omitempty"`
}

type ChartMuseumChartStorageDriverSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// Storage backend name
	Name string `json:"driverName"`

	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

type ChartMuseumChartIndexSpec struct {
	// +kubebuilder:validation:Optional
	// Parallel scan limit for the repo indexer
	ParallelLimit int32 `json:"parallelLimit,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	// Timestamp drift tolerated between cached and generated index before invalidation (in nanoseconds)
	StorageTimestampTolerance time.Duration `json:"storageTimestampTolerance,omitempty"`
}

type ChartMuseumPostFormFieldNameSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="chart"
	// Form field which will be queried for the chart file content
	Chart string `json:"chart,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="prov"
	// Form field which will be queried for the provenance file content
	Provenance string `json:"provenance,omitempty"`
}

type ChartMuseumLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Output structured logs as json
	JSON bool `json:"json"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Show debug messages
	Debug bool `json:"debug"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Log inbound /health requests
	Health bool `json:"health"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// log latency as an integer (nanoseconds) instead of a string
	LatencyInteger bool `json:"latencyInteger"`
}

type ChartMuseumAuthSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Allow anonymous GET operations when auth is used
	AnonymousGet bool `json:"anonymousGet"`

	// +kubebuilder:validation:Optional
	// Reference to secret containing basic http authentication
	BasicAuthRef string `json:"basicAuthRef,omitempty"`

	// +kubebuilder:validation:Optional
	// Bearer authentication specs
	Bearer ChartMuseumAuthBearerSpec `json:"bearer,omitempty"`
}

type ChartMuseumAuthBearerSpec struct {
	// +kubebuilder:validation:Required
	// Reference to secret containing authorization server certificate
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Required
	// Authorization server url
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	// Authorization server service name
	Service string `json:"service"`
}

type ChartMuseumHTTPSSpec struct {
	// +kubebuilder:validation:Required
	// Reference to secret containing tls certificate
	CertificateRef string `json:"certificateRef"`
}

type ChartMuseumDisableSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Disable all routes prefixed with
	API bool `json:"api"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Disable use of index-cache.yaml
	StateFiles bool `json:"statefiles"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Do not allow chart versions to be re-uploaded, even with ?force querystrin
	ForceOverwrite bool `json:"forceOverwrite"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Disable Prometheus metrics
	Metrics bool `json:"metrics"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Disable DELETE route
	Delete bool `json:"delete"`
}

type ChartMuseumCacheSpec struct {
	// +kubebuilder:validation:Optional
	// Redis cache store
	Redis OpacifiedDSN `json:"redis,omitempty"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&ChartMuseum{}, &ChartMuseumList{})
}

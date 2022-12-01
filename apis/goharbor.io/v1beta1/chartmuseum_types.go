package v1beta1

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// ChartMuseum is the Schema for the ChartMuseum API.
type ChartMuseum struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChartMuseumSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
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
	harbormetav1.ComponentSpec `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	Log ChartMuseumLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	Authentication ChartMuseumAuthSpec `json:"authentication,omitempty"`

	// +kubebuilder:validation:Optional
	Server ChartMuseumServerSpec `json:"server,omitempty"`

	// +kubebuilder:validation:Optional
	// Disable some features
	Disable ChartMuseumDisableSpec `json:"disable,omitempty"`

	// +kubebuilder:validation:Optional
	// Cache stores
	Cache ChartMuseumCacheSpec `json:"cache,omitempty"`

	// +kubebuilder:validation:Required
	Chart ChartMuseumChartSpec `json:"chart"`

	// +kubebuilder:validation:Optional
	Network *harbormetav1.Network `json:"network,omitempty"`
}

type ChartMuseumServerSpec struct {
	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Socket timeout
	ReadTimeout *metav1.Duration `json:"readTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Socket timeout
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=20971520
	// Max size of post body (in bytes)
	MaxUploadSize *int64 `json:"maxUploadSize,omitempty"`

	// +kubebuilder:validation:Optional
	// Value to set in the Access-Control-Allow-Origin HTTP header
	CORSAllowOrigin string `json:"corsAllowOrigin,omitempty"`
}

type ChartMuseumChartSpec struct {
	// +kubebuilder:validation:Optional
	// Form fields which will be queried
	PostFormFieldName ChartMuseumPostFormFieldNameSpec `json:"postFormFieldName,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="https?://.*"
	// The absolute url for .tgz files in index.yaml
	URL string `json:"url,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// Allow chart versions to be re-uploaded without ?force querystring
	AllowOverwrite *bool `json:"allowOverwrite,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Enforce the chart museum server only accepts the valid chart version as Helm does
	SemanticVersioning2Only bool `json:"onlySemver2"`

	// +kubebuilder:validation:Required
	Storage ChartMuseumChartStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Index ChartMuseumChartIndexSpec `json:"index,omitempty"`

	// +kubebuilder:validation:Optional
	Repo ChartMuseumChartRepoSpec `json:"repo,omitempty"`
}

type ChartMuseumChartRepoSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The length of repo variable
	DepthDynamic bool `json:"depthDynamic"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// Levels of nested repos for multitenancy
	// Harbor: must be set to 1 to support project namespace
	Depth *int32 `json:"depth,omitempty"`
}

type ChartMuseumChartStorageSpec struct {
	ChartMuseumChartStorageDriverSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// Maximum number of objects allowed in storage (per tenant)
	MaxStorageObjects *int64 `json:"maxStorageObject,omitempty"`
}

type ChartMuseumChartStorageDriverSpec struct {
	// +kubebuilder:validation:Optional
	Amazon *ChartMuseumChartStorageDriverAmazonSpec `json:"amazon,omitempty"`

	// +kubebuilder:validation:Optional
	OpenStack *ChartMuseumChartStorageDriverOpenStackSpec `json:"openstack,omitempty"`

	// +kubebuilder:validation:Optional
	FileSystem *ChartMuseumChartStorageDriverFilesystemSpec `json:"filesystem,omitempty"`

	// +kubebuilder:validation:Optional
	Azure *ChartMuseumChartStorageDriverAzureSpec `json:"azure,omitempty"`

	// +kubebuilder:validation:Optional
	Gcs *ChartMuseumChartStorageDriverGcsSpec `json:"gcs,omitempty"`

	// +kubebuilder:validation:Optional
	Oss *ChartMuseumChartStorageDriverOssSpec `json:"oss,omitempty"`
}

type ChartMuseumChartStorageDriverOssSpec struct {
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// +kubebuilder:validation:Required
	AccessKeyID string `json:"accessKeyID"`

	// +kubebuilder:validation:Required
	AccessSecretRef string `json:"accessSecretRef"`

	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// +kubebuilder:validation:Optional
	PathPrefix string `json:"pathPrefix,omitempty"`
}

type ChartMuseumChartStorageDriverGcsSpec struct {
	// +kubebuilder:validation:Required
	// bucket to store charts for Gcs storage
	Bucket string `json:"bucket"`

	// +kubebuilder:validation:Required
	// The base64 encoded json file which contains the key
	KeyDataSecretRef string `json:"keyDataSecretRef"`

	// +kubebuilder:validation:Optional
	PathPrefix string `json:"pathPrefix,omitempty"`

	// +kubebuilder:validation:Optional
	ChunkSize string `json:"chunksize,omitempty"`
}

type ChartMuseumChartStorageDriverAzureSpec struct {
	// +kubebuilder:validation:Optional
	AccountName string `json:"accountname,omitempty"`

	// +kubebuilder:validation:Optional
	AccountKeyRef string `json:"accountkeyRef,omitempty"`

	// +kubebuilder:validation:Optional
	Container string `json:"container,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=core.windows.net
	BaseURL string `json:"baseURL,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=/azure/harbor/charts
	PathPrefix string `json:"pathPrefix,omitempty"`
}

type ChartMuseumChartStorageDriverAmazonSpec struct {
	// +kubebuilder:validation:Required
	// S3 bucket to store charts for amazon storage
	Bucket string `json:"bucket"`

	// +kubebuilder:validation:Optional
	// Alternative s3 endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// +kubebuilder:validation:Optional
	// Prefix to store charts for the bucket
	Prefix string `json:"prefix,omitempty"`

	// +kubebuilder:validation:Optional
	// Region of the bucket
	Region string `json:"region,omitempty"`

	// +kubebuilder:validation:Optional
	// ServerSideEncryption is the algorithm for server side encryption
	ServerSideEncryption string `json:"serverSideEncryption,omitempty"`

	// +kubebuilder:validation:Optional
	AccessKeyID string `json:"accessKeyID,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	AccessSecretRef string `json:"accessSecretRef,omitempty"`
}

type ChartMuseumChartStorageDriverOpenStackSpec struct {
	// +kubebuilder:validation:Required
	// Container to store charts for openstack storage backend
	Container string `json:"container"`

	// +kubebuilder:validation:Optional
	// Prefix to store charts for the container
	Prefix string `json:"prefix,omitempty"`

	// +kubebuilder:validation:Optional
	// Region of the container
	Region string `json:"region,omitempty"`

	// +kubebuilder:validation:Required
	// URL for obtaining an auth token.
	// https://storage.myprovider.com/v2.0 or https://storage.myprovider.com/v3/auth
	AuthenticationURL string `json:"authenticationURL"`

	// +kubebuilder:validation:Optional
	// Your Openstack tenant name.
	// You can either use tenant or tenantid.
	Tenant string `json:"tenant,omitempty"`

	// +kubebuilder:validation:Optional
	// Your Openstack tenant ID.
	// You can either use tenant or tenantid.
	TenantID string `json:"tenantID,omitempty"`

	// +kubebuilder:validation:Optional
	// Your Openstack domain name for Identity v3 API. You can either use domain or domainid.
	Domain string `json:"domain,omitempty"`

	// +kubebuilder:validation:Optional
	// Your Openstack domain ID for Identity v3 API. You can either use domain or domainid.
	DomainID string `json:"domainID,omitempty"`

	// +kubebuilder:validation:Optional
	// The Openstack user name. You can either use username or userid.
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation:Optional
	// The Openstack user id. You can either use username or userid.
	UserID string `json:"userid,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Secret name containing the Openstack password.
	PasswordRef string `json:"passwordRef,omitempty"`
}

type ChartMuseumChartStorageDriverFilesystemSpec struct {
	// +kubebuilder:validation:Required
	VolumeSource corev1.VolumeSource `json:"volumeSource"`

	// +kubebuilder:validation:Optionel
	Prefix string `json:"prefix,omitempty"`
}

type ChartMuseumChartIndexSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// Parallel scan limit for the repo indexer
	ParallelLimit *int32 `json:"parallelLimit,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Timestamp drift tolerated between cached and generated index before invalidation
	StorageTimestampTolerance *metav1.Duration `json:"storageTimestampTolerance,omitempty"`
}

type ChartMuseumPostFormFieldNameSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default="chart"
	// Form field which will be queried for the chart file content
	// Harbor: Expecting chart to use with Harbor
	Chart string `json:"chart,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default="prov"
	// Form field which will be queried for the provenance file content
	// Harbor: Expecting prov to use with Harbor
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
	// log latency as an integer instead of a string
	LatencyInteger *bool `json:"latencyInteger,omitempty"`
}

type ChartMuseumAuthSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Allow anonymous GET operations when auth is used
	AnonymousGet bool `json:"anonymousGet"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Reference to secret containing basic http authentication
	// Harbor: Harbor try to connect using chart_controller username
	BasicAuthRef string `json:"basicAuthRef,omitempty"`

	// +kubebuilder:validation:Optional
	// Bearer authentication specs
	Bearer *ChartMuseumAuthBearerSpec `json:"bearer,omitempty"`
}

type ChartMuseumAuthBearerSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Reference to secret containing authorization server certificate
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Required
	// Authorization server url
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	// Authorization server service name
	Service string `json:"service"`
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
	Redis *harbormetav1.RedisConnection `json:"redis,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&ChartMuseum{}, &ChartMuseumList{})
}

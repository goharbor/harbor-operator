package v1beta1

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=core
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// Core is the Schema for the Core API.
type Core struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CoreSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CoreList contains a list of Core.
type CoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Core `json:"items"`
}

// CoreSpec defines the desired state of Core.
type CoreSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// https://github.com/goharbor/harbor/blob/master/src/lib/config/metadata/metadatalist.go#L62
	CoreConfig `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	HTTP CoreHTTPSpec `json:"http,omitempty"`

	// +kubebuilder:validation:Required
	Components CoreComponentsSpec `json:"components"`

	// +kubebuilder:validation:Optional
	Proxy *harbormetav1.ProxySpec `json:"proxy,omitempty"`

	// +kubebuilder:validation:Optional
	Log CoreLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Required
	Database CoreDatabaseSpec `json:"database"`

	// +kubebuilder:validation:Required
	Redis CoreRedisSpec `json:"redis"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	ExternalEndpoint string `json:"externalEndpoint"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	ConfigExpiration *metav1.Duration `json:"configExpiration,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CSRFKeyRef string `json:"csrfKeyRef"`

	// +kubebuilder:validation:Optional
	Metrics *harbormetav1.MetricsSpec `json:"metrics,omitempty"`

	// +kubebuilder:validation:Optional
	Network *harbormetav1.Network `json:"network,omitempty"`

	// +kubebuilder:validation:Optional
	Trace *harbormetav1.TraceSpec `json:"trace,omitempty"`
}

type CoreRedisSpec struct {
	harbormetav1.RedisConnection `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="30s"
	// IdleTimeoutSecond closes connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty"`
}

type CoreHTTPSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	GZip *bool `json:"enableGzip,omitempty"`
}

type CoreComponentsSpec struct {
	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Required
	JobService CoreComponentsJobServiceSpec `json:"jobService"`

	// +kubebuilder:validation:Required
	Portal CoreComponentPortalSpec `json:"portal"`

	// +kubebuilder:validation:Required
	Registry CoreComponentsRegistrySpec `json:"registry"`

	// +kubebuilder:validation:Required
	TokenService CoreComponentsTokenServiceSpec `json:"tokenService"`

	// +kubebuilder:validation:Optional
	Trivy *CoreComponentsTrivySpec `json:"trivy,omitempty"`

	// +kubebuilder:validation:Optional
	ChartRepository *CoreComponentsChartRepositorySpec `json:"chartRepository,omitempty"`

	// +kubebuilder:validation:Optional
	NotaryServer *CoreComponentsNotaryServerSpec `json:"notaryServer,omitempty"`
}

type CoreComponentPortalSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`
}

const (
	CoreDatabaseType = "postgresql"
)

type CoreDatabaseSpec struct {
	harbormetav1.PostgresConnectionWithParameters `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=50
	MaxIdleConnections *int32 `json:"maxIdleConnections,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1000
	MaxOpenConnections *int32 `json:"maxOpenConnections,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	EncryptionKeyRef string `json:"encryptionKeyRef"`
}

type CoreComponentsJobServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	SecretRef string `json:"secretRef"`
}

type CoreComponentsRegistrySpec struct {
	RegistryControllerConnectionSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	Redis *harbormetav1.RedisConnection `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Sync bool `json:"sync"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	StorageProviderName string `json:"storageProviderName,omitempty"`
}

type RegistryControllerConnectionSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	RegistryURL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	ControllerURL string `json:"controllerURL"`

	// +kubebuilder:validation:Required
	Credentials CoreComponentsRegistryCredentialsSpec `json:"credentials"`
}

type CoreComponentsRegistryCredentialsSpec struct {
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	PasswordRef string `json:"passwordRef"`
}

type CoreComponentsTokenServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef"`
}

type CoreComponentsChartRepositorySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	AbsoluteURL bool `json:"absoluteURL"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"redis"}
	// +kubebuilder:default="redis"
	CacheDriver string `json:"cacheDriver,omitempty"`
}

type CoreComponentsTrivySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	AdapterURL string `json:"adapterURL"`
}

type CoreComponentsNotaryServerSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`
}

type CoreConfig struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	AdminInitialPasswordRef string `json:"adminInitialPasswordRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"db_auth"}
	// +kubebuilder:default="db_auth"
	AuthenticationMode string `json:"authMode,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	SecretRef string `json:"secretRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	PublicCertificateRef string `json:"publicCertificateRef,omitempty"`
}

type CoreLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	Level harbormetav1.CoreLogLevel `json:"level,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Core{}, &CoreList{})
}

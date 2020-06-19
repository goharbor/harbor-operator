package v1alpha2

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RegistryCorePublicURLKey = "REGISTRY_HTTP_HOST"
	RegistryAuthURLKey       = "REGISTRY_AUTH_TOKEN_REALM" // RegistryCorePublicURLKey + "/service/token"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=registry
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// Registry is the Schema for the registries API.
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec RegistrySpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// RegistryList contains a list of Registry.
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Registry `json:"items"`
}

// RegistrySpec defines the desired state of Registry.
// See https://docs.docker.com/registry/configuration/
type RegistrySpec struct {
	ComponentSpec    `json:",inline"`
	RegistryConfig01 `json:",inline"`
}

type RegistryConfig01 struct {
	// +kubebuilder:validation:Optional
	Log RegistryLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	HTTP RegistryHTTPSpec `json:"http,omitempty"`

	// +kubebuilder:validation:Optional
	Health RegistryHealthSpec `json:"health,omitempty"`

	// +kubebuilder:validation:Optional
	Notifications RegistryNotificationsSpec `json:"notifications,omitempty"`

	// +kubebuilder:validation:Optional
	Authentication RegistryAuthenticationSpec `json:"authentication,omitempty"`

	// +kubebuilder:validation:Optional
	Validation RegistryValidationSpec `json:"validation,omitempty"`

	// +kubebuilder:validation:Optional
	Compatibility RegistryCompatibilitySpec `json:"compatibility,omitempty"`

	// +kubebuilder:validation:Required
	Storage RegistryStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Middlewares RegistryMiddlewaresSpec `json:"middlewares,omitempty"`

	// +kubebuilder:validation:Optional
	Reporting map[string]string `json:"reporting,omitempty"`

	// +kubebuilder:validation:Optional
	Redis RegistryRedisSpec `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	Proxy RegistryProxySpec `json:"proxy,omitempty"`
}

type RegistryProxySpec struct {
	// +kubebuilder:validation:Required
	RemoteURL string `json:"remoteURL"`

	// +kubebuilder:validation:Optional
	BasicAuthRef string `json:"basicAuthRef,omitempty"`
}

type RegistryRedisSpec struct {
	OpacifiedDSN `json:",inline"`

	// +kubebuilder:validation:Optional
	DialTimeout time.Duration `json:"dialTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	ReadTimeout time.Duration `json:"readTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	Pool RegistryRedisPoolSpec `json:"pool,omitempty"`
}

type RegistryRedisPoolSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=3
	MaxIdle int `json:"maxIdle,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5
	MaxActive int `json:"maxActive,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30000000000
	IdleTimeout time.Duration `json:"idleTimeout,omitempty"`
}

type RegistryLogSpec struct {
	// +kubebuilder:validation:Optional
	AccessLog RegistryAccessLogSpec `json:"accessLog,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"debug","info","warn","error"}
	// +kubebuilder:default="info"
	Level string `json:"level,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"text","json","logstash"}
	// +kubebuilder:default="text"
	Formatter string `json:"formatter,omitempty"`

	// +kubebuilder:validation:Optional
	Fields map[string]string `json:"fields,omitempty"`

	// +kubebuilder:validation:Optional
	Hooks []RegistryLogHookSpec `json:"hooks,omitempty"`
}

type RegistryAccessLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Disabled bool `json:"disabled"`
}

type RegistryLogHookSpec struct {
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// +kubebuilder:validation:Required
	Levels []string `json:"levels"`

	// +kubebuilder:validation:Required
	OptionsRef string `json:"optionsRef"`
}

type RegistryMiddlewaresSpec struct {
	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	Registry []RegistryMiddlewareSpec `json:"registry,omitempty"`

	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	Repository []RegistryMiddlewareSpec `json:"repository,omitempty"`

	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	Storage []RegistryMiddlewareSpec `json:"storage,omitempty"`
}

type RegistryMiddlewareSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	OptionsRef string `json:"optionsRef,omitempty"`
}

type RegistryHTTPSpec struct {
	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`

	// +kubebuilder:validation:Optional
	Host string `json:"host,omitempty"`

	// +kubebuilder:validation:Optional
	Net string `json:"net,omitempty"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix,omitempty"`

	// +kubebuilder:validation:Optional
	DrainTimeout time.Duration `json:"drainTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	Headers map[string][]string `json:"headers,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	RelativeURLs bool `json:"relativeURLs"`

	// +kubebuilder:validation:Optional
	HTTP2 RegistryHTTPHTTP2Spec `json:"http2,omitempty"`

	// +kubebuilder:validation:Optional
	Debug RegistryHTTPDebugSpec `json:"debug,omitempty"`
}

type RegistryHTTPHTTP2Spec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Disabled bool `json:"disabled"`
}

type RegistryHTTPDebugSpec struct {
	// +kubebuilder:validation:Required
	Address string `json:"address"`

	// +kubebuilder:validation:Optional
	Prometheus RegistryHTTPDebugPrometheusSpec `json:"prometheus,omitempty"`
}

type RegistryHTTPDebugPrometheusSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/metrics"
	Path string `json:"path,omitempty"`
}

type RegistryHealthSpec struct {
	// +kubebuilder:validation:Optional
	StorageDriver RegistryHealthStorageDriverSpec `json:"storageDriver,omitempty"`

	// +kubebuilder:validation:Optional
	File []RegistryHealthFileSpec `json:"file,omitempty"`

	// +kubebuilder:validation:Optional
	HTTP []RegistryHealthHTTPSpec `json:"http,omitempty"`

	// +kubebuilder:validation:Optional
	TCP []RegistryHealthTCPSpec `json:"tcp,omitempty"`
}

type RegistryHealthStorageDriverSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	Interval time.Duration `json:"interval,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=3
	Threshold int `json:"threshold,omitempty"`
}

type RegistryHealthFileSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	File string `json:"path"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	Interval time.Duration `json:"interval,omitempty"`
}

type RegistryHealthHTTPSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URI string `json:"uri"`

	// +kubebuilder:validation:Optional
	Headers map[string][]string `json:"headers,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=200
	StatusCode int `json:"statuscode,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	Timeout time.Duration `json:"timeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	Interval time.Duration `json:"interval,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=3
	Threshold int `json:"threshold,omitempty"`
}

type RegistryHealthTCPSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Address string `json:"address"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	Timeout time.Duration `json:"timeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	Interval time.Duration `json:"interval,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=3
	Threshold int `json:"threshold,omitempty"`
}

type RegistryNotificationsSpec struct {
	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	Endpoints []RegistryNotificationEndpointSpec `json:"endpoints,omitempty"`

	// +kubebuilder:validation:Optional
	Events RegistryNotificationEventsSpec `json:"events,omitempty"`
}

type RegistryNotificationEventsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	IncludeReferences bool `json:"includeReferences"`
}

type RegistryNotificationEndpointSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Disabled bool `json:"disabled"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	Headers map[string][]string `json:"headers"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=5000000000
	Timeout time.Duration `json:"timeout"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=3
	Threshold int `json:"threshold"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=10000000000
	Backoff time.Duration `json:"backoff"`

	// +kubebuilder:validation:Optional
	IgnoredMediaTypes []string `json:"ignoredMediaTypes,omitempty"`

	// +kubebuilder:validation:Optional
	Ignore RegistryNotificationEndpointIgnoreSpec `json:"ignore,omitempty"`
}

type RegistryNotificationEndpointIgnoreSpec struct {
	// +kubebuilder:validation:Optional
	MediaTypes []string `json:"mediaTypes,omitempty"`

	// +kubebuilder:validation:Optional
	Actions []string `json:"actions,omitempty"`
}

type RegistryAuthenticationSpec struct {
	// +kubebuilder:validation:Optional
	Silly RegistryAuthenticationSillySpec `json:"silly,omitempty"`

	// +kubebuilder:validation:Optional
	Token RegistryAuthenticationTokenSpec `json:"token,omitempty"`

	// +kubebuilder:validation:Optional
	HTPasswd RegistryAuthenticationHTPasswdSpec `json:"htPasswd,omitempty"`
}

type RegistryAuthenticationSillySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Service string `json:"service"`
}

type RegistryAuthenticationTokenSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Service string `json:"service"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Issuer string `json:"issuer"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:true
	AutoRedirect bool `json:"autoredirect"`
}

type RegistryAuthenticationHTPasswdSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	SecretRef string `json:"secretRef"`
}

type RegistryValidationSpec struct {
	// +kubebuilder:validation:Optional
	Disabled bool `json:"disabled"`

	// +kubebuilder:validation:Optional
	Manifests RegistryValidationManifestSpec `json:"manifests,omitempty"`
}

type RegistryValidationManifestSpec struct {
	// +kubebuilder:validation:Optional
	URLs RegistryValidationManifestURLsSpec `json:"urls,omitempty"`
}

type RegistryValidationManifestURLsSpec struct {
	// +kubebuilder:validation:Optional
	Allow []string `json:"allow,omitempty"`

	// +kubebuilder:validation:Optional
	Deny []string `json:"deny,omitempty"`
}

type RegistryCompatibilitySpec struct {
	// +kubebuilder:validation:Optional
	Schema1 RegistryCompatibilitySchemaSpec `json:"schema1,omitempty"`
}

type RegistryCompatibilitySchemaSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	CertificateRef string `json:"certificateRef,omitempty"`
}

type RegistryStorageSpec struct {
	RegistryStorageDriverSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	Cache RegistryStorageCacheSpec `json:"cache"`

	// +kubebuilder:validation:Optional
	Maintenance RegistryStorageMaintenanceSpec `json:"maintenance,omitempty"`

	// +kubebuilder:validation:Optional
	Delete RegistryStorageDeleteSpec `json:"delete,omitempty"`

	// +kubebuilder:validation:Optional
	Redirect RegistryStorageRedirectSpec `json:"redirect,omitempty"`
}

type RegistryStorageDriverSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"driverName"`

	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef"`
}

type RegistryStorageRedirectSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Disable bool `json:"disable"`
}

type RegistryStorageCacheSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"inmemory","redis"}
	Blobdescriptor string `json:"blobdescriptor"`
}

type RegistryStorageMaintenanceSpec struct {
	// +kubebuilder:validation:Optional
	UploadPurging RegistryStorageMaintenanceUploadPurgingSpec `json:"uploadPurging,omitempty"`

	// +kubebuilder:validation:Optional
	ReadOnly RegistryStorageMaintenanceReadOnlySpec `json:"readOnly,omitempty"`
}

type RegistryStorageMaintenanceReadOnlySpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`
}

type RegistryStorageMaintenanceUploadPurgingSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DryRun bool `json:"dryRun"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=3600000000000
	Age time.Duration `json:"age,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=900000000000
	Interval time.Duration `json:"interval,omitempty"`
}

type RegistryStorageDeleteSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}

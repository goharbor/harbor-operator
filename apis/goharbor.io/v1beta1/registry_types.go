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
// +resource:path=registry
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// Registry is the Schema for the registries API.
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec RegistrySpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
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
	harbormetav1.ComponentSpec `json:",inline"`
	RegistryConfig01           `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	Proxy *harbormetav1.ProxySpec `json:"proxy,omitempty"`

	// +kubebuilder:validation:Optional
	Network *harbormetav1.Network `json:"network,omitempty"`

	// +kubebuilder:validation:Optional
	Trace *harbormetav1.TraceSpec `json:"trace,omitempty"`
}

func (r *RegistrySpec) Default() {
	if r.Storage.Cache.Blobdescriptor == "" {
		if r.Redis == nil {
			r.Storage.Cache.Blobdescriptor = "inmemory"
		} else {
			r.Storage.Cache.Blobdescriptor = "redis"
		}
	}
}

type RegistryConfig01 struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"level":"info","formatter":"text"}
	Log RegistryLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	HTTP RegistryHTTPSpec `json:"http,omitempty"`

	// +kubebuilder:validation:Optional
	Health RegistryHealthSpec `json:"health,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
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
	Redis *RegistryRedisSpec `json:"redis,omitempty"`
}

type RegistryRedisSpec struct {
	harbormetav1.RedisConnection `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	DialTimeout *metav1.Duration `json:"dialTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	ReadTimeout *metav1.Duration `json:"readTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	Pool RegistryRedisPoolSpec `json:"pool,omitempty"`
}

type RegistryRedisPoolSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	MaxIdle *int32 `json:"maxIdle,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=5
	MaxActive *int32 `json:"maxActive,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="30s"
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty"`
}

type RegistryLogSpec struct {
	// +kubebuilder:validation:Optional
	AccessLog RegistryAccessLogSpec `json:"accessLog,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	Level harbormetav1.RegistryLogLevel `json:"level,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="text"
	Formatter harbormetav1.RegistryLogFormatter `json:"formatter,omitempty"`

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
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
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
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	OptionsRef string `json:"optionsRef,omitempty"`
}

type RegistryHTTPSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// The secret name containing a random piece of data
	// used to sign state that may be stored with the client
	// to protect against tampering. For production environments
	// you should generate a random piece of data using
	// a cryptographically secure random generator.
	// If you omit the secret, the registry will automatically generate a secret when it starts.
	// If you are building a cluster of registries behind a load balancer,
	// you MUST ensure the secret is the same for all registries.
	SecretRef string `json:"secretRef,omitempty"`

	// +kubebuilder:validation:Optional
	// A fully-qualified URL for an externally-reachable address for the registry.
	// If present, it is used when creating generated URLs.
	// Otherwise, these URLs are derived from client requests.
	Host string `json:"host,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"unix","tcp"}
	// +kubebuilder:default="tcp"
	// The network used to create a listening socket.
	Net string `json:"net,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="(/(.+/)?)?"
	// If the server does not run at the root path, set this to the value of the prefix.
	// The root path is the section before v2.
	// It requires both preceding and trailing slashes, such as in the example /path/.
	Prefix string `json:"prefix,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Amount of time to wait for HTTP connections to drain
	// before shutting down after registry receives SIGTERM signal
	DrainTimeout *metav1.Duration `json:"drainTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"X-Content-Type-Options":{"nosniff"}}
	// Use this option to specify headers that the HTTP server should include in responses.
	// This can be used for security headers such as Strict-Transport-Security.
	// The headers option should contain an option for each header to include, where the parameter
	// name is the header’s name, and the parameter value a list of the header’s payload values.
	// Including X-Content-Type-Options: [nosniff] is recommended, sothat browsers
	// will not interpret content as HTML if they are directed to load a page from the registry.
	// This header is included in the example configuration file.
	Headers map[string][]string `json:"headers,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// If true, the registry returns relative URLs in Location headers.
	// The client is responsible for resolving the correct URL.
	// This option is not compatible with Docker 1.7 and earlier.
	RelativeURLs *bool `json:"relativeURLs,omitempty"`

	// +kubebuilder:validation:Optional
	// Use the http2 structure to control http2 settings for the registry.
	HTTP2 RegistryHTTPHTTP2Spec `json:"http2,omitempty"`

	// +kubebuilder:validation:Optional
	// Use debug option to configure a debug server that can be helpful in diagnosing problems.
	// The debug endpoint can be used for monitoring registry metrics and health,
	// as well as profiling. Sensitive information may be available via the debug endpoint.
	// Please be certain that access to the debug endpoint is locked down in a production environment.
	Debug *RegistryHTTPDebugSpec `json:"debug,omitempty"`

	// +kubebuilder:validation:Optional
	// Use this to configure TLS for the server.
	// If you already have a web server running on the same host as the registry,
	// you may prefer to configure TLS on that web server and proxy connections to the registry server.
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`
}

type RegistryHTTPHTTP2Spec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Disabled bool `json:"disabled"`
}

type RegistryHTTPDebugSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:default=5001
	Port int32 `json:"port,omitempty"`

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
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:default=3
	Threshold int32 `json:"threshold,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	Interval *metav1.Duration `json:"interval,omitempty"`
}

type RegistryHealthFileSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	File string `json:"path"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	Interval *metav1.Duration `json:"interval,omitempty"`
}

type RegistryHealthHTTPSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URI string `json:"uri"`

	// +kubebuilder:validation:Optional
	Headers map[string][]string `json:"headers,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	Interval *metav1.Duration `json:"interval,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	Threshold *int32 `json:"threshold,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=200
	StatusCode *int32 `json:"statuscode,omitempty"`
}

type RegistryHealthTCPSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Address string `json:"address"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	Interval *metav1.Duration `json:"interval,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	Threshold *int32 `json:"threshold,omitempty"`
}

type RegistryNotificationsSpec struct {
	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	// The endpoints structure contains a list of named services (URLs) that can accept event notifications.
	Endpoints []RegistryNotificationEndpointSpec `json:"endpoints,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Events RegistryNotificationEventsSpec `json:"events,omitempty"`
}

type RegistryNotificationEventsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	IncludeReferences *bool `json:"includeReferences,omitempty"`
}

type RegistryNotificationEndpointSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	// A human-readable name for the service.
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	// The URL to which events should be published.
	URL string `json:"url"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// If true, notifications are disabled for the service.
	Disabled bool `json:"disabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	Threshold *int32 `json:"threshold,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="5s"
	// A value for the HTTP timeout. A positive integer and an optional suffix indicating the unit of time, which may be ns, us, ms, s, m, or h. If you omit the unit of time, ns is used.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="10s"
	Backoff *metav1.Duration `json:"backoff,omitempty"`

	// +kubebuilder:validation:Optional
	Headers map[string][]string `json:"headers,omitempty"`

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
	Silly *RegistryAuthenticationSillySpec `json:"silly,omitempty"`

	// +kubebuilder:validation:Optional
	Token *RegistryAuthenticationTokenSpec `json:"token,omitempty"`

	// +kubebuilder:validation:Optional
	HTPasswd *RegistryAuthenticationHTPasswdSpec `json:"htPasswd,omitempty"`
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
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:true
	AutoRedirect *bool `json:"autoredirect,omitempty"`
}

type RegistryAuthenticationHTPasswdSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:minLength=1
	Realm string `json:"realm"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
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
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef,omitempty"`
}

type RegistryStorageSpec struct {
	// +kubebuilder:validation:Required
	Driver RegistryStorageDriverSpec `json:"driver"`

	// +kubebuilder:validation:Optional
	Cache RegistryStorageCacheSpec `json:"cache,omitempty"`

	// +kubebuilder:validation:Optional
	Maintenance RegistryStorageMaintenanceSpec `json:"maintenance,omitempty"`

	// +kubebuilder:validation:Optional
	Delete RegistryStorageDeleteSpec `json:"delete,omitempty"`

	// +kubebuilder:validation:Optional
	Redirect RegistryStorageRedirectSpec `json:"redirect,omitempty"`
}

type RegistryStorageDriverSpec struct {
	// +kubebuilder:validation:Optional
	// InMemory storage driver is for purely tests purposes.
	// This driver is an implementation of the storagedriver.StorageDriver interface which
	// uses local memory for object storage.
	// If you would like to run a registry from volatile memory, use the filesystem driver on a ramdisk.
	// IMPORTANT: This storage driver does not persist data across runs. This is why it is only suitable for testing. Never use this driver in production.
	// See: https://docs.docker.com/registry/storage-drivers/inmemory/
	InMemory *RegistryStorageDriverInmemorySpec `json:"inmemory,omitempty"`

	// +kubebuilder:validation:Optional
	// FileSystem is an implementation of the storagedriver.StorageDriver interface which uses the local filesystem.
	// The local filesystem can be a remote volume.
	// See: https://docs.docker.com/registry/storage-drivers/filesystem/
	FileSystem *RegistryStorageDriverFilesystemSpec `json:"filesystem,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Amazon S3 or S3 compatible services for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/s3/
	S3 *RegistryStorageDriverS3Spec `json:"s3,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface that uses OpenStack Swift for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/swift/
	Swift *RegistryStorageDriverSwiftSpec `json:"swift,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Microsoft Azure Blob Storage for object storage.
	// See: https://docs.docker.com/registry/storage-drivers/azure/
	Azure *RegistryStorageDriverAzureSpec `json:"azure,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Google Cloud for object storage.
	// https://docs.docker.com/registry/storage-drivers/gcs/
	Gcs *RegistryStorageDriverGcsSpec `json:"gcs,omitempty"`

	// +kubebuilder:validation:Optional
	// An implementation of the storagedriver.StorageDriver interface which uses Alibaba Cloud for object storage.
	// https://docs.docker.com/registry/storage-drivers/oss/
	Oss *RegistryStorageDriverOssSpec `json:"oss,omitempty"`
}

func (r *RegistryStorageDriverSpec) Validate() error {
	found := 0

	if r.InMemory != nil {
		found++
	}

	if r.FileSystem != nil {
		found++
	}

	if r.S3 != nil {
		found++
	}

	if r.Swift != nil {
		found++
	}

	if r.Azure != nil {
		found++
	}

	if r.Gcs != nil {
		found++
	}

	if r.Oss != nil {
		found++
	}

	switch found {
	case 0:
		return ErrNoStorageConfiguration
	case 1:
		return nil
	default:
		return Err2StorageConfiguration
	}
}

type RegistryStorageDriverInmemorySpec struct{}

type RegistryStorageDriverFilesystemSpec struct {
	// +kubebuilder:validation:Required
	VolumeSource corev1.VolumeSource `json:"volumeSource"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=25
	// +kubebuilder:default=100
	MaxThreads int32 `json:"maxthreads,omitempty"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix,omitempty"`
}

type RegistryStorageDriverAzureSpec struct {
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

type RegistryStorageDriverOssSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="oss-.*"
	Region string `json:"region"`

	// +kubebuilder:validation:Required
	AccessKeyID string `json:"accessKeyID"`

	// +kubebuilder:validation:Required
	AccessSecretRef string `json:"accessSecretRef"`

	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// +kubebuilder:validation:Optional
	PathPrefix string `json:"pathPrefix,omitempty"`

	// +kubebuilder:validation:Optional
	Endpoint string `json:"endpoint,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Internal bool `json:"internal,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Specifies whether the registry stores the image in encrypted format or not. A boolean value.
	Encrypt bool `json:"encrypt,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Secure *bool `json:"secure,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=5242880
	// The Oss API requires multipart upload chunks to be at least 5MB.
	ChunkSize int64 `json:"chunksize,omitempty"`
}

type RegistryStorageDriverGcsSpec struct {
	// +kubebuilder:validation:Required
	// The base64 encoded json file which contains the key
	KeyDataRef string `json:"keyDataRef,omitempty"`

	// +kubebuilder:validation:Required
	// bucket to store charts for Gcs storage
	Bucket string `json:"bucket,omitempty"`

	// +kubebuilder:validation:Optional
	PathPrefix string `json:"pathPrefix,omitempty"`

	// +kubebuilder:validation:Optional
	ChunkSize string `json:"chunkSize,omitempty"`
}

type RegistryStorageDriverS3Spec struct {
	// +kubebuilder:validation:Optional
	// The AWS Access Key.
	// If you use IAM roles, omit to fetch temporary credentials from IAM.
	AccessKey string `json:"accesskey,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Reference to the secret containing the AWS Secret Key.
	// If you use IAM roles, omit to fetch temporary credentials from IAM.
	SecretKeyRef string `json:"secretkeyRef,omitempty"`

	// +kubebuilder:validation:Required
	// The AWS region in which your bucket exists.
	// For the moment, the Go AWS library in use does not use the newer DNS based bucket routing.
	// For a list of regions, see http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
	Region string `json:"region"`

	// +kubebuilder:validation:Optional
	// Endpoint for S3 compatible storage services (Minio, etc).
	RegionEndpoint string `json:"regionendpoint,omitempty"`

	// +kubebuilder:validation:Required
	// The bucket name in which you want to store the registry’s data.
	Bucket string `json:"bucket"`

	// +kubebuilder:validation:Optional
	// This is a prefix that is applied to all S3 keys to allow you to segment data in your bucket if necessary.
	RootDirectory string `json:"rootdirectory,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="STANDARD"
	// The S3 storage class applied to each registry file.
	StorageClass string `json:"storageclass,omitempty"`

	// +kubebuilder:validation:Optional
	// KMS key ID to use for encryption (encrypt must be true, or this parameter is ignored).
	KeyID string `json:"keyid,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Specifies whether the registry stores the image in encrypted format or not. A boolean value.
	Encrypt bool `json:"encrypt"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Skips TLS verification when the value is set to true.
	SkipVerify bool `json:"skipverify"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Secure *bool `json:"secure,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// Indicates whether the registry uses Version 4 of AWS’s authentication.
	V4Auth *bool `json:"v4auth,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=5242880
	// The S3 API requires multipart upload chunks to be at least 5MB.
	ChunkSize int64 `json:"chunksize,omitempty"`

	// +kubebuilder:validation:Optional
	MultipartCopyChunkSize int64 `json:"multipartcopychunksize,omitempty"`

	// +kubebuilder:validation:Optional
	MultipartCopyMaxConcurrency int64 `json:"multipartcopymaxconcurrency,omitempty"`

	// +kubebuilder:validation:Optional
	MultipartCopyThresholdSize int64 `json:"multipartcopythresholdsize,omitempty"`
}

type RegistryStorageDriverSwiftSpec struct {
	// +kubebuilder:validation:Required
	// URL for obtaining an auth token.
	// https://storage.myprovider.com/v2.0 or https://storage.myprovider.com/v3/auth
	AuthURL string `json:"authurl"`

	// +kubebuilder:validation:Required
	// The Openstack user name.
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Secret name containing the Openstack password.
	PasswordRef string `json:"passwordRef,omitempty"`

	// +kubebuilder:validation:Optional
	// The Openstack region in which your container exists.
	Region string `json:"region,omitempty"`

	// +kubebuilder:validation:Required
	// The name of your Swift container where you wish to store the registry’s data.
	// The driver creates the named container during its initialization.
	Container string `json:"container"`

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
	// Your Openstack trust ID for Identity v3 API.
	TrustID string `json:"trustid,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Skips TLS verification if the value is set to true.
	InsecureSkipVerify bool `json:"insecureskipverify,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=5242880
	// Size of the data segments for the Swift Dynamic Large Objects.
	// This value should be a number.
	ChunkSize int64 `json:"chunksize,omitempty"`

	// +kubebuilder:validation:Optional
	// This is a prefix that is applied to all Swift keys to allow you to segment data in your container if necessary. Defaults to the container’s root.
	Prefix string `json:"prefix,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// The secret key used to generate temporary URLs.
	SecretKeyRef string `json:"secretkeyRef,omitempty"`

	// +kubebuilder:validation:Optional
	// The access key to generate temporary URLs. It is used by HP Cloud Object Storage in addition to the secretkey parameter.
	AccessKey string `json:"accesskey,omitempty"`

	// +kubebuilder:validation:Optional
	// Specify the OpenStack Auth’s version, for example 3. By default the driver autodetects the auth’s version from the authurl.
	AuthVersion string `json:"authversion,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="public"
	// +kubebuilder:validation:Enum={"public","internal","admin"}
	// The endpoint type used when connecting to swift.
	EndpointType string `json:"endpointtype,omitempty"`
}

type RegistryStorageRedirectSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Disable bool `json:"disable"`
}

type RegistryStorageCacheSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"inmemory","redis"}
	Blobdescriptor string `json:"blobdescriptor,omitempty"`
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
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="168h"
	Age *metav1.Duration `json:"age,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// +kubebuilder:default="24h"
	Interval *metav1.Duration `json:"interval,omitempty"`
}

type RegistryStorageDeleteSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}

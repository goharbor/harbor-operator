package v1alpha2

import (
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
	// +kubebuilder:validation:Required
	Driver RegistryStorageDriverSpec `json:"driver"`

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
}

var (
	errNoStorageConfiguration = errors.New("no storage configuration")
	err2StorageConfiguration  = errors.New("only 1 storage can be configured")
)

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

	switch found {
	case 0:
		return errNoStorageConfiguration
	case 1:
		return nil
	default:
		return err2StorageConfiguration
	}
}

type RegistryStorageDriverInmemorySpec struct{}

type RegistryStorageDriverFilesystemSpec struct {
	// +kubebuilder:validation:Required
	VolumeSource corev1.VolumeSource `json:"volumeSource"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=25
	// +kubebuilder:default=100
	MaxThreads int `json:"maxthreads"`
}

type RegistryStorageDriverS3Spec struct {
	// +kubebuilder:validation:Optional
	// The AWS Access Key.
	// If you use IAM roles, omit to fetch temporary credentials from IAM.
	AccessKey string `json:"accesskey,omitempty"`

	// +kubebuilder:validation:Optional
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
	// +kubebuilder:default=true
	Secure bool `json:"secure"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Skips TLS verification when the value is set to true.
	SkipVerify bool `json:"skipverify"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// Indicates whether the registry uses Version 4 of AWS’s authentication.
	V4Auth bool `json:"v4auth"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=5242880
	// The S3 API requires multipart upload chunks to be at least 5MB.
	ChunkSize int64 `json:"chunksize,omitempty"`
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
	// The secret key used to generate temporary URLs.
	SecretKeyRef string `json:"secretkeyRef,omitempty"`

	// +kubebuilder:validation:Optional
	// The access key to generate temporary URLs. It is used by HP Cloud Object Storage in addition to the secretkey parameter.
	AccessKey string `json:"accesskey,omitempty"`

	// +kubebuilder:validation:Optional
	// Specify the OpenStack Auth’s version, for example 3. By default the driver autodetects the auth’s version from the AuthURL.
	AuthVersion string `json:"authversion,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="public"
	// +kubebuilder:validation:Enum={"public","internal","admin"}
	// The endpoint type used when connecting to swift.
	EndpointType string `json:"endpointtype"`
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

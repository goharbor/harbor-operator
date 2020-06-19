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
// +resource:path=core
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// Core is the Schema for the registries API.
type Core struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CoreSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
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
	ComponentSpec `json:",inline"`
	// https://github.com/goharbor/harbor/blob/master/src/common/config/metadata/metadatalist.go#L62
	CoreConfig `json:",inline"`

	// +kubebuilder:validation:Optional
	HTTP CoreHTTPSpec `json:"http"`

	// +kubebuilder:validation:Required
	Components CoreComponentsSpec `json:"components"`

	// +kubebuilder:validation:Optional
	Proxy CoreProxySpec `json:"proxy,omitempty"`

	// +kubebuilder:validation:Optional
	Log CoreLogSpec `json:"log"`

	// +kubebuilder:validation:Required
	Database CoreDatabaseSpec `json:"database"`

	// +kubebuilder:validation:Required
	Redis CoreRedisSpec `json:"redis"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	ExternalEndpoint string `json:"externalEndpoint"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5000000000
	ConfigExpiration time.Duration `json:"configExpiration"`

	// +kubebuilder:validation:Required
	CSRFKeyRef string `json:"csrfKeyRef"`
}

type CoreRedisSpec struct {
	OpacifiedDSN `json:",inline"`

	// IdleTimeoutSecond closes connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30000000000
	IdleTimeout time.Duration `json:"idleTimeout"`
}

type CoreHTTPSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	GZip bool `json:"enableGzip"`
}

type CoreComponentsSpec struct {
	// +kubebuilder:validation:Optional
	TLS CoreComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Required
	JobService CoreComponentsJobServiceSpec `json:"jobService"`

	// +kubebuilder:validation:Required
	Portal CoreComponentPortalSpec `json:"portal"`

	// +kubebuilder:validation:Required
	Registry CoreComponentsRegistrySpec `json:"registry"`

	// +kubebuilder:validation:Required
	TokenService CoreComponentsTokenServiceSpec `json:"tokenService"`

	// +kubebuilder:validation:Optional
	Trivy CoreComponentsTrivySpec `json:"trivy,omitempty"`

	// +kubebuilder:validation:Optional
	Clair CoreComponentsClairSpec `json:"clairAdapter,omitempty"`

	// +kubebuilder:validation:Optional
	ChartRepository CoreComponentsChartRepositorySpec `json:"chartRepository,omitempty"`

	// +kubebuilder:validation:Optional
	NotaryServer CoreComponentsNotaryServerSpec `json:"notaryServer,omitempty"`
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
	CorePostgresqlSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MaxIdleConnections int `json:"maxIdleConnections"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MaxOpenConnections int `json:"maxOpenConnections"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	EncryptionKeyRef string `json:"encryptionKeyRef"`
}

type CorePostgresqlSpec struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:default=5432
	Port int `json:"port"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default="postgres"
	Username string `json:"username"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default="postgres"
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	PasswordRef string `json:"passwordRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"disable","allow","prefer","require","verify-ca","verify-full"}
	// +kubebuilder:default="prefer"
	SSLMode string `json:"sslMode"`
}

type CoreComponentsJobServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`
}

type CoreComponentsRegistrySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	ControllerURL string `json:"controllerURL"`

	// +kubebuilder:validation:Optional
	Redis OpacifiedDSN `json:"redis"`

	// +kubebuilder:validation:Required
	Credentials CoreComponentsRegistryCredentialsSpec `json:"credentials"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Sync bool `json:"sync"`

	// +kubebuilder:validation:Optional
	StorageProviderName string `json:"storageProviderName"`
}

type CoreComponentsRegistryCredentialsSpec struct {
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// +kubebuilder:validation:Required
	PasswordRef string `json:"passwordRef"`
}

type CoreComponentsTokenServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`
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
	CacheDriver string `json:"cacheDriver"`
}

type CoreComponentsTLSSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Verify bool `json:"verify"`

	// +kubebuilder:validation:Optional
	CertificateRef string `json:"certificateRef,omitempty"`
}

type CoreComponentsTrivySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	AdapterURL string `json:"adapterURL"`
}

type CoreComponentsClairSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	AdapterURL string `json:"adapterURL"`

	// +kubebuilder:validation:Required
	Database CorePostgresqlSpec `json:"database"`
}

type CoreComponentsNotaryServerSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`
}

type CoreConfig struct {
	// +kubebuilder:validation:Required
	AdminInitialPasswordRef string `json:"adminInitialPasswordRef"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"db_auth"}
	// +kubebuilder:default="db_auth"
	AuthenticationMode string `json:"authMode"`

	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`

	// +kubebuilder:validation:Optional
	PublicCertificateRef string `json:"publicCertificateRef,omitempty"`
}

type CoreProxySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	URL string `json:"url"`

	// +kubebuilder:validation:Optional
	NoProxy []string `json:"noProxy"`

	// +kubebuilder:validation:Optional
	Components []string `json:"components"`
}

type CoreLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"debug","info","warn","error"}
	// +kubebuilder:default="info"
	Level string `json:"level"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Core{}, &CoreList{})
}

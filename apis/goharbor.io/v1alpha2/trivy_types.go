package v1alpha2

import (
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=trivy
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// Trivy is the Schema for the Trivy API.
type Trivy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TrivySpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TrivyList contains a list of Trivy.
type TrivyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Trivy `json:"items"`
}

// TrivySpec defines the desired state of Trivy.
type TrivySpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	Log TrivyLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	Server TrivyServerSpec `json:"server,omitempty"`

	// +kubebuilder:validation:Required
	// Cache stores
	Cache TrivyCacheSpec `json:"cache,omitempty"`
}

type TrivyServerSpec struct {
	// +kubebuilder:validation:Optional
	HTTPS TrivyHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="15s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Socket timeout
	ReadTimeout *metav1.Duration `json:"readTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="15s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Socket timeout
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="60s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// Idle timeout
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=":8080"
	// +kubebuilder:validation:Pattern=".*:[0-9]{0,5}"
	// Binding address for the API server
	Address string `json:"address,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/home/scanner/.cache/trivy"
	// Trivy cache directory
	CacheDir string `json:"cacheDir,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/home/scanner/.cache/reports"
	// Trivy reports directory
	ReportsDir string `json:"reportsDir,omitempty"`

	// +kubebuilder:validation:Optional
	// Comma-separated list of vulnerability types.
	// Possible values are os and library.
	VulnType []TrivyServerVulnerabilityType `json:"vulnType,omitempty"`

	// +kubebuilder:validation:Optional
	// Comma-separated list of vulnerabilities
	// severities to be displayed
	Severity []TrivyServerSeverityType `json:"severity,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to display only fixed vulnerabilities
	IgnoreUnfixed bool `json:"ignoreUnfixed,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to enable or disable Trivy debug mode
	DebugMode bool `json:"debugMode,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to enable or disable Trivy DB downloads from GitHub
	SkipUpdate bool `json:"skipUpdate,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to skip verifying registry certificate
	Insecure bool `json:"insecure,omitempty"`

	// +kubebuilder:validation:Optional
	// The GitHub access token to download Trivy DB (see GitHub rate limiting)
	GithubToken string `json:"githubToken,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="http?://.+"
	// The URL of the HTTP proxy server
	HTTProxy string `json:"httpProxy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="https?://.+"
	// The URL of the HTTPS proxy server
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// +kubebuilder:validation:Optional
	// The URLs that the proxy settings do not apply to
	NoProxy []string `json:"noProxy,omitempty"`
}

// +kubebuilder:validation:Enum={"os","library"}
// +kubebuilder:default={"os","library"}
// TrivyServerVulnerabilityType represents a CVE vulnerability type for trivy.
type TrivyServerVulnerabilityType string

// +kubebuilder:validation:Enum={"UNKNOWN","LOW","MEDIUM","HIGH","CRITICAL"}
// +kubebuilder:default={"UNKNOWN","LOW","MEDIUM","HIGH","CRITICAL"}
// TrivyServerSeverityType represents a CVE severity type for trivy.
type TrivyServerSeverityType string

var trivyURLValidationRegexp = regexp.MustCompile(`https?://.+`)

func (r *TrivyServerSpec) Validate() map[string]error {
	errors := map[string]error{}

	if len(r.NoProxy) > 0 {
		for _, url := range r.NoProxy {
			matched := trivyURLValidationRegexp.MatchString(url)
			if !matched {
				errors["NoProxy"] = ErrWrongURLFormat
				break
			}
		}
	}

	return errors
}

type TrivyLogSpec struct {
	// +kubebuilder:validation:Optional
	Level harbormetav1.TrivyLogLevel `json:"level,omitempty"`
}

type TrivyHTTPSSpec struct {
	// +kubebuilder:validation:Optionnal
	// Reference to secret containing tls certificate
	CertificateRef string `json:"certificateRef"`

	// +kubebuilder:validation:Optionnal
	// Reference to secret containing tls key
	KeyRef string `json:"keyRef"`

	// +kubebuilder:validation:Optionnal
	// A list of absolute paths to x509 root certificate authorities
	// that the api use if required to verify a client certificate
	ClientCas string `json:"clientCasList"`
}

type TrivyCacheSpec struct {
	// +kubebuilder:validation:Required
	// Redis cache store
	Redis OpacifiedDSN `json:"redis,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="harbor.scanner.trivy:store"
	// The namespace for keys in the Redis store
	RedisNamespace string `json:"redisNamespace,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1h"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The time to live for persisting scan jobs and associated scan reports
	RedisScanJobTTL *metav1.Duration `json:"redisScanJobTTL,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="harbor.scanner.trivy:job-queue"
	// The namespace for keys in the scan jobs queue backed by Redis
	QueueRedisNamespace string `json:"queueRedisNamespace,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// The number of workers to spin-up for the scan jobs queue
	QueueWorkerConcurrency int `json:"queueWorkerConcurrency,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5
	// The max number of connections allocated by the Redis connection pool
	PoolMaxActive int `json:"poolMaxActive,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5
	// The max number of idle connections in the Redis connection pool
	PoolMaxIdle int `json:"poolMaxIdle,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="5m"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The duration after which idle connections to the Redis server are closed.
	// If the value is zero, then idle connections are not closed.
	PoolIdleTimeout *metav1.Duration `json:"poolIdleTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The timeout for connecting to the Redis server
	PoolConnectionTimeout *metav1.Duration `json:"poolConnectionTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The timeout for reading a single Redis command reply
	PoolReadTimeout *metav1.Duration `json:"poolReadTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The timeout for writing a single Redis command
	PoolWriteTimeout *metav1.Duration `json:"poolWriteTimeout,omitempty"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Trivy{}, &TrivyList{})
}

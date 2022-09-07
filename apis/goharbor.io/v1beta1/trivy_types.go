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
// +resource:path=trivy
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
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

	harbormetav1.TrivyVulnerabilityTypes `json:",inline"`

	harbormetav1.TrivySeverityTypes `json:",inline"`

	CertificateInjection `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"level":"info"}
	Log TrivyLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Required
	Server TrivyServerSpec `json:"server"`

	// +kubebuilder:validation:Optional
	Update TrivyUpdateSpec `json:"update,omitempty"`

	// +kubebuilder:validation:Required
	// Redis cache store
	Redis TrivyRedisSpec `json:"redis,omitempty"`

	// +kubebuilder:validation:Required
	Storage TrivyStorageSpec `json:"storage"`

	// +kubebuilder:validation:Optional
	Proxy *harbormetav1.ProxySpec `json:"proxy,omitempty"`

	// +kubebuilder:validation:Optional
	Network *harbormetav1.Network `json:"network,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="5m0s"
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	OfflineScan bool `json:"offlineScan"`
}

type TrivyUpdateSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to enable or disable Trivy DB downloads from GitHub
	Skip bool `json:"skip"`

	// +kubebuilder:validation:Optional
	// The GitHub access token to download Trivy DB (see GitHub rate limiting)
	GithubTokenRef string `json:"githubTokenRef,omitempty"`
}

type TrivyServerSpec struct {
	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	ClientCertificateAuthorityRefs []string `json:"clientCertificateAuthorityRefs,omitempty"`

	// +kubebuilder:validation:Optional
	TokenServiceCertificateAuthorityRefs []string `json:"tokenServiceCertificateAuthorityRefs,omitempty"`

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
	// +kubebuilder:default=false
	// The flag to display only fixed vulnerabilities
	IgnoreUnfixed bool `json:"ignoreUnfixed,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to enable or disable Trivy debug mode
	DebugMode bool `json:"debugMode,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// The flag to skip verifying registry certificate
	Insecure bool `json:"insecure,omitempty"`

	// +kubebuilder:validation:Optional
	Proxy *TrivyServerProxySpec `json:"proxy,omitempty"`
}

type TrivyServerProxySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	// The URL of the proxy server
	URL string `json:"URL"`

	// +kubebuilder:validation:Optional
	// The URLs that the proxy settings do not apply to
	NoProxy []string `json:"noProxy,omitempty"`
}

type TrivyLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	Level harbormetav1.TrivyLogLevel `json:"level,omitempty"`
}

type TrivyRedisSpec struct {
	harbormetav1.RedisConnection `json:",inline"`

	// +kubebuilder:validation:Required
	Pool TrivyRedisPoolSpec `json:"pool,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="harbor.scanner.trivy:store"
	// The namespace for keys in the Redis store
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:Optional
	Jobs TrivyRedisJobsSpec `json:"jobs,omitempty"`
}

type TrivyRedisJobsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1h"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The time to live for persisting scan jobs and associated scan reports
	ScanTTL *metav1.Duration `json:"scanTTL,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="harbor.scanner.trivy:job-queue"
	// The namespace for keys in the scan jobs queue backed by Redis
	Namespace string `json:"Namespace,omitempty"`
}

type TrivyRedisPoolSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=0
	// The max number of connections allocated by the Redis connection pool
	MaxActive int `json:"maxActive,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=0
	// The max number of idle connections in the Redis connection pool
	MaxIdle int `json:"maxIdle,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="5m"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The duration after which idle connections to the Redis server are closed.
	// If the value is zero, then idle connections are not closed.
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The timeout for connecting to the Redis server
	ConnectionTimeout *metav1.Duration `json:"connectionTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The timeout for reading a single Redis command reply
	ReadTimeout *metav1.Duration `json:"readTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default="1s"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
	// The timeout for writing a single Redis command
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`
}

type TrivyStorageSpec struct {
	// +kubebuilder:validation:Required
	Reports TrivyStorageVolumeSpec `json:"reports"`

	// +kubebuilder:validation:Required
	Cache TrivyStorageVolumeSpec `json:"cache"`
}

type TrivyStorageVolumeSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"emptyDir":{"sizeLimit":"1Gi"}}
	VolumeSource corev1.VolumeSource `json:"volumeSource,omitempty"`

	// +kubebuilder:validation:Optional
	Prefix string `json:"prefix,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Trivy{}, &TrivyList{})
}

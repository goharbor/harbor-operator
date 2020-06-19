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
// +resource:path=jobservice
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The semver Harbor version",priority=5
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.spec.replicas`,description="The number of replicas",priority=0
// JobService is the Schema for the registries API.
type JobService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JobServiceSpec `json:"spec,omitempty"`

	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// JobServiceList contains a list of JobService.
type JobServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobService `json:"items"`
}

// JobServiceSpec defines the desired state of JobService.
type JobServiceSpec struct {
	ComponentSpec       `json:",inline"`
	JobServiceComponent `json:",inline"`

	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`

	// Config to use https protocol
	// +kubebuilder:validation:Optional
	HTTPS JobServiceHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Required
	Core JobServiceCoreSpec `json:"core"`

	// Configurations of worker pool
	// +kubebuilder:validation:Required
	WorkerPool JobServicePoolSpec `json:"workerPool,omitempty"`

	// Job logger configurations
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	JobLoggers []JobServiceLoggerConfigSpec `json:"jobLoggers"`

	// Logger configurations
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Loggers []JobServiceLoggerConfigSpec `json:"loggers"`

	// +kubebuilder:validation:Required
	Registry CoreComponentsRegistryCredentialsSpec `json:"registry"`
}

type JobServiceCoreSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	SecretRef string `json:"secretRef"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=".+://.+"
	URL string `json:"url"`
}

type JobServiceHTTPSSpec struct {
	// +kubebuilder:validation:Required
	CertificateRef string `json:"certificateRef,omitempty"`
}

// RedisPoolConfig keeps redis worker info.
type JobServicePoolRedisSpec struct {
	OpacifiedDSN `json:",inline"`

	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// IdleTimeoutSecond closes connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30000000000
	IdleTimeout time.Duration `json:"idleTimeout"`
}

// PoolConfig keeps worker worker configurations.
type JobServicePoolSpec struct {
	// Worker concurrency
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	WorkerCount uint `json:"workers"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"redis"}
	// +kubebuilder:default="redis"
	Backend string `json:"backend"`
	// +kubebuilder:validation:Optional
	Redis JobServicePoolRedisSpec `json:"redisPool,omitempty"`
}

// JobServiceLoggerConfigSweeperSpec keeps settings of log sweeper.
type JobServiceLoggerConfigSweeperSpec struct {
	// +kubebuilder:validation:Optional
	Duration int `json:"duration"`
	// +kubebuilder:validation:Optional
	SettingsRef string `json:"settingsRef"`
}

// LoggerConfig keeps logger basic configurations.
type JobServiceLoggerConfigSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"DB","FILE","STD_OUTPUT"}
	Name string `json:"name"` // https://github.com/goharbor/harbor/blob/master/src/jobservice/logger/known_loggers.go#L9-L16
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"DEBUG","INFO","WARNING","ERROR","FATAL"}
	// +kubebuilder:default=INFO
	Level string `json:"level"`
	// +kubebuilder:validation:Optional
	SettingsRef string `json:"settingsRef"`
	// +kubebuilder:validation:Optional
	Sweeper *JobServiceLoggerConfigSweeperSpec `json:"sweeper"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&JobService{}, &JobServiceList{})
}

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
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
	ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	SecretRef string `json:"secretRef"`

	// +kubebuilder:validation:Optional
	// Config to use https protocol
	HTTPS JobServiceHTTPSSpec `json:"https,omitempty"`

	// +kubebuilder:validation:Required
	Core JobServiceCoreSpec `json:"core"`

	// +kubebuilder:validation:Required
	// Configurations of worker pool
	WorkerPool JobServicePoolSpec `json:"workerPool,omitempty"`

	// +kubebuilder:validation:Required
	// Job logger configurations
	JobLoggers JobServiceLoggerConfigSpec `json:"jobLoggers,omitempty"`

	// +kubebuilder:validation:Required
	// Logger configurations
	Loggers JobServiceLoggerConfigSpec `json:"loggers,omitempty"`

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

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="30s"
	// IdleTimeoutSecond closes connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout PositiveDuration `json:"idleTimeout,omitempty"`
}

// PoolConfig keeps worker worker configurations.
type JobServicePoolSpec struct {
	// Worker concurrency
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10
	WorkerCount int32 `json:"workers,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Enum={"redis"}
	// +kubebuilder:default="redis"
	Backend string `json:"backend,omitempty"`

	// +kubebuilder:validation:Optional
	Redis JobServicePoolRedisSpec `json:"redisPool,omitempty"`
}

// JobServiceLoggerConfigSweeperSpec keeps settings of log sweeper.
type JobServiceLoggerConfigSweeperSpec struct {

	// +kubebuilder:validation:Optional
	SettingsRef string `json:"settingsRef,omitempty"`
}

// LoggerConfig keeps logger basic configurations.
// One of files, database or stdout must be defined.
type JobServiceLoggerConfigSpec struct {
	// +kubebuilder:validation:Optional
	// +nullable
	Files []JobServiceLoggerConfigFileSpec `json:"files,omitempty"`

	// +kubebuilder:validation:Optional
	Database *JobServiceLoggerConfigDatabaseSpec `json:"database,omitempty"`

	// +kubebuilder:validation:Optional
	// +nullable
	STDOUT *JobServiceLoggerConfigSTDOUTSpec `json:"stdout,omitempty"`
}

type JobServiceLoggerConfigDatabaseSpec struct {
	// +kubebuilder:validation:Required
	Level JobServiceLogLevel `json:"level"`

	// +kubebuilder:validation:Optional
	Sweeper PositiveDuration `json:"sweeper,omitempty"`
}

type JobServiceLoggerConfigSTDOUTSpec struct {
	// +kubebuilder:validation:Required
	Level JobServiceLogLevel `json:"level"`
}

type JobServiceLoggerConfigFileSpec struct {
	// +kubebuilder:validation:Optional
	Volume *corev1.VolumeSource `json:"volume,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default="INFO"
	Level JobServiceLogLevel `json:"level"`

	// +kubebuilder:validation:Optional
	Sweeper PositiveDuration `json:"sweeper,omitempty"`
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"DB","FILE","STD_OUTPUT"}
// JobServiceLoggerName is the type of logger to configure.
type JobServiceLoggerName string

const (
	JobServiceLoggerDatabase JobServiceLoggerName = "DB"
	JobServiceLoggerFile     JobServiceLoggerName = "FILE"
	JobServiceLoggerSTDOUT   JobServiceLoggerName = "STD_OUTPUT"
)

// +kubebuilder:validation:Type=string

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&JobService{}, &JobServiceList{})
}

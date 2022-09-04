package v1beta1

import (
	"strconv"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=exporter
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",priority=1
// +kubebuilder:printcolumn:name="Failure",type=string,JSONPath=`.status.conditions[?(@.type=="Failed")].message`,description="Human readable message describing the failure",priority=5
// Exporter is the Schema for the Exporter API.
type Exporter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ExporterSpec `json:"spec,omitempty"`

	Status harbormetav1.ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ExporterList contains a list of Exporter.
type ExporterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Exporter `json:"items"`
}

// ExporterSpec defines the desired state of Exporter.
type ExporterSpec struct {
	harbormetav1.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	TLS *harbormetav1.ComponentsTLSSpec `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	Log ExporterLogSpec `json:"log,omitempty"`

	// +kubebuilder:validation:Optional
	Cache ExporterCacheSpec `json:"cache,omitempty"`

	// +kubebuilder:validation:Required
	Core ExporterCoreSpec `json:"core"`

	// +kubebuilder:validation:Required
	Database ExporterDatabaseSpec `json:"database"`

	// +kubebuilder:validation:Optional
	JobService *ExporterJobServiceSpec `json:"jobservice,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=8001
	// +kubebuilder:validation:Minimum=1
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/metrics"
	// +kubebuilder:validation:Pattern="/.+"
	Path string `json:"path,omitempty"`

	// +kubebuilder:validation:Optional
	Network *harbormetav1.Network `json:"network,omitempty"`
}

type ExporterCacheSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?"
	// +kubebuilder:default="30s"
	// The duration to cache info from the database and core.
	Duration *metav1.Duration `json:"duration,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?"
	// +kubebuilder:default="4h"
	// The interval to clean the cache info from the database and core.
	CleanInterval *metav1.Duration `json:"cleanInterval,omitempty"`
}

const (
	defaultCacheDurationSeconds = 30

	defaultCacheCleanIntervalSeconds = 4 * 60 * 60 // 4hours

	baseInt10 = 10
)

func (spec *ExporterCacheSpec) GetDurationEnvVar() string {
	seconds := int64(defaultCacheDurationSeconds)
	if spec.Duration != nil {
		seconds = int64(spec.Duration.Seconds())
	}

	return strconv.FormatInt(seconds, baseInt10)
}

func (spec *ExporterCacheSpec) GetCleanIntervalEnvVar() string {
	seconds := int64(defaultCacheCleanIntervalSeconds)
	if spec.CleanInterval != nil {
		seconds = int64(spec.CleanInterval.Seconds())
	}

	return strconv.FormatInt(seconds, baseInt10)
}

type ExporterCoreSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="https?://.+"
	// The absolute Harbor Core URL.
	URL string `json:"url"`
}

type ExporterDatabaseSpec struct {
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

type ExporterJobServiceSpec struct {
	// +kubebuilder:validation:Optional
	Redis *JobServicePoolRedisSpec `json:"redisPool,omitempty"`
}

type ExporterLogSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	Level harbormetav1.ExporterLogLevel `json:"level,omitempty"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Exporter{}, &ExporterList{})
}

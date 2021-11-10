package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/goharbor/harbor-operator/pkg/config/harbor"
	"github.com/goharbor/harbor/src/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// JaegerAgentPasswordKey is the password to connect to jaeger.
	JaegerAgentPasswordKey = "jaeger-password"
)

// +kubebuilder:validation:Enum={"jaeger", "otel"}
// +kubebuilder:validation:Type="string"
// The tracing provider: 'jaeger' or 'otel'.
type TraceProviderType string

const (
	TraceJaegerProvider TraceProviderType = "jaeger"
	TraceOtelProvider   TraceProviderType = "otel"
)

type TraceSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// Enable tracing or not.
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Optional
	// Namespace used to differentiate different harbor services.
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// Set `sampleRate` to 1 if you wanna sampling 100% of trace data; set 0.5 if you wanna sampling 50% of trace data, and so forth.
	SampleRate int `json:"sampleRate,omitempty"`

	// +kubebuilder:validation:Optional
	// A key value dict contains user defined attributes used to initialize trace provider.
	Attributes map[string]string `json:"attributes,omitempty"`

	// +kubebuilder:validation:Required
	Provder TraceProviderType `json:"provider"`

	TraceProviderSpec `json:",inline"`
}

func (spec *TraceSpec) Validate(rootPath *field.Path) *field.Error {
	if !spec.IsEnabled() {
		return nil
	}

	if rootPath == nil {
		rootPath = field.NewPath("spec").Child("trace")
	}

	switch spec.Provder {
	case TraceJaegerProvider:
		if spec.Jaeger == nil {
			return field.Required(rootPath.Child("jaeger"), fmt.Sprintf("field is required for %s provider", spec.Provder))
		}

		if err := spec.Jaeger.Validate(rootPath.Child("jaeger")); err != nil {
			return err
		}
	case TraceOtelProvider:
		if spec.Otel == nil {
			if spec.Jaeger == nil {
				return field.Required(rootPath.Child("otel"), fmt.Sprintf("field is required for %s provider", spec.Provder))
			}
		}
	}

	return nil
}

func (spec *TraceSpec) IsEnabled() bool {
	return spec != nil && spec.Enabled
}

func (spec *TraceSpec) AddEnvVars(component string, envs []corev1.EnvVar) ([]corev1.EnvVar, error) {
	traceEnvs, err := spec.GetEnvVars(component)
	if err != nil {
		return nil, err
	}

	return append(envs, traceEnvs...), nil
}

func (spec *TraceSpec) GetEnvVars(component string) ([]corev1.EnvVar, error) {
	if !spec.IsEnabled() {
		return harbor.EnvVars(map[string]harbor.ConfigValue{
			common.TraceEnabled: harbor.Value(strconv.FormatBool(false)),
		})
	}

	configs := map[string]harbor.ConfigValue{
		common.TraceEnabled:     harbor.Value(strconv.FormatBool(spec.Enabled)),
		common.TraceSampleRate:  harbor.Value(strconv.Itoa(spec.SampleRate)),
		common.TraceNamespace:   harbor.Value(spec.Namespace),
		common.TraceServiceName: harbor.Value(fmt.Sprintf("harbor-%s", component)),
	}

	if len(spec.Attributes) > 0 {
		attrs, _ := json.Marshal(spec.Attributes)
		configs[common.TraceAttributes] = harbor.Value(string(attrs))
	}

	switch spec.Provder {
	case TraceJaegerProvider:
		for k, v := range spec.Jaeger.getConfigValues() {
			configs[k] = v
		}
	case TraceOtelProvider:
		for k, v := range spec.Otel.getConfigValues() {
			configs[k] = v
		}
	}

	return harbor.EnvVars(configs)
}

type TraceProviderSpec struct {
	// +kubebuilder:validation:Optional
	Jaeger *JaegerSpec `json:"jaeger,omitempty"`

	// +kubebuilder:validation:Optional
	Otel *OtelSpec `json:"otel,omitempty"`
}

// +kubebuilder:validation:Enum={"collector", "agent"}
// +kubebuilder:validation:Type="string"
// The jaeger mode: 'collector' or 'agent'.
type JaegerModeType string

const (
	JaegerCollectorMode JaegerModeType = "collector"
	JaegerAgentMode     JaegerModeType = "agent"
)

type JaegerSpec struct {
	// +kubebuilder:validation:Required
	Mode JaegerModeType `json:"mode"`

	// +kubebuilder:validation:Optional
	Collector *JaegerCollectorSpec `json:"collector,omitempty"`

	// +kubebuilder:validation:Optional
	Agent *JaegerAgentSpec `json:"agent,omitempty"`
}

func (spec *JaegerSpec) Validate(rootPath *field.Path) *field.Error {
	if rootPath == nil {
		rootPath = field.NewPath("jaeger")
	}

	switch spec.Mode {
	case JaegerCollectorMode:
		if spec.Collector == nil {
			return field.Required(rootPath.Child("collector"), fmt.Sprintf("field is required in %s mode", spec.Mode))
		}
	case JaegerAgentMode:
		if spec.Agent == nil {
			return field.Required(rootPath.Child("agent"), fmt.Sprintf("field is required in %s mode", spec.Mode))
		}
	}

	return nil
}

func (spec *JaegerSpec) getConfigValues() map[string]harbor.ConfigValue {
	switch spec.Mode {
	case JaegerCollectorMode:
		return spec.Collector.getConfigValues()
	case JaegerAgentMode:
		return spec.Agent.getConfigValues()
	default:
		return nil
	}
}

type JaegerCollectorSpec struct {
	// +kubebuilder:validation:Required
	// The endpoint of the jaeger collector.
	Endpoint string `json:"endpoint"`

	// +kubebuilder:validation:Optional
	// The username of the jaeger collector.
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation:Optional
	// The password secret reference name of the jaeger collector.
	PasswordRef string `json:"passwordRef,omitempty"`
}

func (spec *JaegerCollectorSpec) getConfigValues() map[string]harbor.ConfigValue {
	configs := map[string]harbor.ConfigValue{
		common.TraceJaegerEndpoint: harbor.Value(spec.Endpoint),
		common.TraceJaegerUsername: harbor.Value(spec.Username),
		common.TraceJaegerPassword: harbor.Value(""),
	}

	if spec.PasswordRef != "" {
		configs[common.TraceJaegerPassword] = harbor.ValueFrom{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: PostgresqlPasswordKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: spec.PasswordRef,
				},
			},
		}
	}

	return configs
}

type JaegerAgentSpec struct {
	// +kubebuilder:validation:Required
	// The host of the jaeger agent.
	Host string `json:"host,omitempty"`

	// +kubebuilder:validation:Required
	// The port of the jaeger agent.
	Port int `json:"port,omitempty"`
}

func (spec *JaegerAgentSpec) getConfigValues() map[string]harbor.ConfigValue {
	return map[string]harbor.ConfigValue{
		common.TraceJaegerAgentHost: harbor.Value(spec.Host),
		common.TraceJaegerAgentPort: harbor.Value(strconv.Itoa(spec.Port)),
	}
}

type OtelSpec struct {
	// +kubebuilder:validation:Required
	// The endpoint of otel.
	Endpoint string `json:"endpoint"`

	// +kubebuilder:validation:Required
	// The URL path of otel.
	URLPath string `json:"urlPath"`

	// +kubebuilder:validation:Optional
	// Whether enable compression or not for otel.
	Compression bool `json:"compression,omitempty"`

	// +kubebuilder:validation:Optional
	// Whether establish insecure connection or not for otel.
	Insecure bool `json:"insecure,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="10s"
	// The timeout of otel.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

func (spec *OtelSpec) getConfigValues() map[string]harbor.ConfigValue {
	return map[string]harbor.ConfigValue{
		common.TraceOtelEndpoint:    harbor.Value(spec.Endpoint),
		common.TraceOtelURLPath:     harbor.Value(spec.URLPath),
		common.TraceOtelCompression: harbor.Value(strconv.FormatBool(spec.Compression)),
		common.TraceOtelInsecure:    harbor.Value(strconv.FormatBool(spec.Insecure)),
		common.TraceOtelTimeout:     harbor.Value(fmt.Sprintf("%d", int64(spec.Timeout.Seconds()))),
	}
}

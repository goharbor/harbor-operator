package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type ProxySpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="https?://.+"
	HTTPProxy string `json:"httpProxy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="https?://.+"
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"127.0.0.1","localhost",".local",".internal"}
	NoProxy []string `json:"noProxy,omitempty"`
}

func (spec *ProxySpec) GetEnvVars() []corev1.EnvVar {
	if spec == nil {
		return nil
	}

	var envVars []corev1.EnvVar

	if spec.HTTPProxy != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "HTTP_PROXY",
			Value: spec.HTTPProxy,
		})
	}

	if spec.HTTPSProxy != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "HTTPS_PROXY",
			Value: spec.HTTPSProxy,
		})
	}

	if len(spec.NoProxy) > 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "NO_PROXY",
			Value: strings.Join(spec.NoProxy, ","),
		})
	}

	return envVars
}

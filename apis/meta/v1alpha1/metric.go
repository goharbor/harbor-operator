package v1alpha1

import (
	"fmt"
)

const (
	PrometheusScrapeAnnotationKey = "prometheus.io/scrape"
	PrometheusPathAnnotationKey   = "prometheus.io/path"
	PrometheusPortAnnotationKey   = "prometheus.io/port"
)

type MetricsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=8001
	// +kubebuilder:validation:Minimum=1
	// The port of the metrics.
	Port int32 `json:"port"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/metrics"
	// +kubebuilder:validation:Pattern="/.+"
	// The path of the metrics.
	Path string `json:"path"`
}

func (spec *MetricsSpec) IsEnabled() bool {
	return spec != nil && spec.Enabled
}

func (spec *MetricsSpec) AddPrometheusAnnotations(annotations map[string]string) map[string]string {
	if !spec.IsEnabled() {
		return annotations
	}

	return AddPrometheusAnnotations(annotations, spec.Port, spec.Path)
}

func AddPrometheusAnnotations(annotations map[string]string, port int32, path string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[PrometheusScrapeAnnotationKey] = "true"
	annotations[PrometheusPathAnnotationKey] = path
	annotations[PrometheusPortAnnotationKey] = fmt.Sprintf("%d", port)

	return annotations
}

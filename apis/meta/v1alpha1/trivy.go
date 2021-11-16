package v1alpha1

import "strings"

type TrivyVulnerabilityTypes struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"os","library"}
	// Comma-separated list of vulnerability types.
	VulnerabilityTypes *[]TrivyVulnerabilityType `json:"vulnerabilityTypes,omitempty"`
}

// GetVulnerabilities joins array of vulnerabilities type into a string separated by commas.
func (v *TrivyVulnerabilityTypes) GetValue() string {
	if v.VulnerabilityTypes == nil {
		return ""
	}

	vulnerabilities := make([]string, len(*v.VulnerabilityTypes))

	for index, v := range *v.VulnerabilityTypes {
		vulnerabilities[index] = string(v)
	}

	return strings.Join(vulnerabilities, ",")
}

// +kubebuilder:validation:Enum={"os","library"}
// +kubebuilder:validation:Type="string"
// TrivyServerVulnerabilityType represents a CVE vulnerability type for trivy.
type TrivyVulnerabilityType string

const (
	TrivyOSVulnerability      TrivyVulnerabilityType = "os"
	TrivyLibraryVulnerability TrivyVulnerabilityType = "library"
)

type TrivySeverityTypes struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"UNKNOWN","LOW","MEDIUM","HIGH","CRITICAL"}
	// List of severities to be displayed
	Severities *[]TrivySeverityType `json:"severities,omitempty"`
}

// GetSeverities joins array of severities type into a string separated by commas.
func (s *TrivySeverityTypes) GetValue() string {
	if s.Severities == nil {
		return ""
	}

	severities := make([]string, len(*s.Severities))

	for index, v := range *s.Severities {
		severities[index] = string(v)
	}

	return strings.Join(severities, ",")
}

// +kubebuilder:validation:Enum={"UNKNOWN","LOW","MEDIUM","HIGH","CRITICAL"}
// +kubebuilder:validation:Type="string"
// TrivyServerSeverityType represents a CVE severity type for trivy.
type TrivySeverityType string

const (
	TrivyUnknownSeverity  TrivySeverityType = "UNKNOWN"
	TrivyLowSeverity      TrivySeverityType = "LOW"
	TrivyMediumSeverity   TrivySeverityType = "MEDIUM"
	TrivyHighSeverity     TrivySeverityType = "HIGH"
	TrivyCriticalSeverity TrivySeverityType = "CRITICAL"
)

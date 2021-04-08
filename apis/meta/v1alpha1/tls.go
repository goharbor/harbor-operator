package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type ComponentWithTLS Component

const (
	CoreTLS               = ComponentWithTLS(CoreComponent)
	ChartMuseumTLS        = ComponentWithTLS(ChartMuseumComponent)
	ExporterTLS           = ComponentWithTLS(ExporterComponent)
	JobServiceTLS         = ComponentWithTLS(JobServiceComponent)
	PortalTLS             = ComponentWithTLS(PortalComponent)
	RegistryTLS           = ComponentWithTLS(RegistryComponent)
	RegistryControllerTLS = ComponentWithTLS(RegistryControllerComponent)
	NotaryServerTLS       = ComponentWithTLS(NotaryServerComponent)
	TrivyTLS              = ComponentWithTLS(TrivyComponent)
)

func (r ComponentWithTLS) String() string {
	return Component(r).String()
}

func (r ComponentWithTLS) GetName() string {
	return r.String()
}

type ComponentsTLSSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef,omitempty"`
}

func (tls *ComponentsTLSSpec) Enabled() bool {
	return tls != nil
}

func (tls *ComponentsTLSSpec) GetScheme() corev1.URIScheme {
	if tls.Enabled() {
		return corev1.URISchemeHTTPS
	}

	return corev1.URISchemeHTTP
}

const (
	HTTPSPort = 443
	HTTPPort  = 80
)

func (tls *ComponentsTLSSpec) GetInternalPort() int32 {
	if tls.Enabled() {
		return HTTPSPort
	}

	return HTTPPort
}

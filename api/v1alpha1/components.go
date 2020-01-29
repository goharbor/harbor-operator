package v1alpha1

import "fmt"

const (
	CoreName        = "core"
	RegistryName    = "registry"
	CertificateName = "certificate"
	JobServiceName  = "jobservice"
	PortalName      = "portal"
	NotaryName      = "notary"
	ClairName       = "clair"
	ChartMuseumName = "chartmuseum"
)

func (h *Harbor) NormalizeComponentName(componentName string) string {
	return fmt.Sprintf("%s-%s", h.GetName(), componentName)
}

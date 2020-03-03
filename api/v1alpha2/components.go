package v1alpha2

import "fmt"

const (
	CertificateName  = "certificate"
	ChartMuseumName  = "chartmuseum"
	ClairName        = "clair"
	CoreName         = "core"
	HarborName       = "harbor"
	JobServiceName   = "jobservice"
	NotaryName       = "notary"
	NotaryServerName = "notary-server"
	NotarySignerName = "notary-signer"
	PortalName       = "portal"
	RegistryName     = "registry"
)

func (h *Harbor) NormalizeComponentName(componentName string) string {
	return fmt.Sprintf("%s-%s", h.GetName(), componentName)
}

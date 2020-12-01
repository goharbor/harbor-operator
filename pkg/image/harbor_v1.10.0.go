package image

type harborV1_10_0ImageLocator struct {
}

func (h harborV1_10_0ImageLocator) CoreImage() string {
	return "goharbor/harbor-core:v1.10.0"
}

func (h harborV1_10_0ImageLocator) ChartMuseumImage() string {
	return "goharbor/chartmuseum-photon:v0.9.0-v1.10.0"
}

func (h harborV1_10_0ImageLocator) ClairImage() string {
	return "goharbor/clair-photon:v2.1.1-v1.10.0"
}

func (h harborV1_10_0ImageLocator) ClairAdapterImage() string {
	// As it mentioned, https://github.com/goharbor/harbor-operator/blob/44ab8a074b3ebda2c94d29268a7fc823c9fe97a9/api/v1alpha1/harbor_image.go#L29
	// Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	return "holyhope/clair-adapter-with-config:v1.10.0"
}

func (h harborV1_10_0ImageLocator) JobServiceImage() string {
	return "goharbor/harbor-jobservice:v1.10.0"
}

func (h harborV1_10_0ImageLocator) NotaryServerImage() string {
	return "goharbor/notary-server-photon:v0.6.1-v1.10.0"
}

func (h harborV1_10_0ImageLocator) NotarySingerImage() string {
	return "goharbor/notary-signer-photon:v0.6.1-v1.10.0"
}

func (h harborV1_10_0ImageLocator) NotaryDBMigratorImage() string {
	return "jmonsinjon/notary-db-migrator:v0.6.1"
}

func (h harborV1_10_0ImageLocator) PortalImage() string {
	return "goharbor/harbor-portal:v1.10.0"
}

func (h harborV1_10_0ImageLocator) RegistryImage() string {
	return "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.10.0"
}

func (h harborV1_10_0ImageLocator) RegistryControllerImage() string {
	return "goharbor/harbor-registryctl:v1.10.0"
}

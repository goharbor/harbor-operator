package v1alpha1

func (component *CoreComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/harbor-core:v1.10.0"
	}

	return *component.Image
}

func (component *ChartMuseumComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/chartmuseum-photon:v0.9.0-v1.10.0"
	}

	return *component.Image
}

func (component *ClairComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/clair-photon:v2.1.1-v1.10.0"
	}

	return *component.Image
}

func (component *ClairAdapterComponent) GetImage() string {
	if component.Image == nil {
		return "holyhope/clair-adapter-with-config:v1.10.0" // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	}

	return *component.Image
}

func (component *JobServiceComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/harbor-jobservice:v1.10.0"
	}

	return *component.Image
}
func (component *NotaryServerComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/notary-server-photon:v0.6.1-v1.10.0"
	}

	return *component.Image
}

func (component *NotarySignerComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/notary-signer-photon:v0.6.1-v1.10.0"
	}

	return *component.Image
}
func (component *NotaryDBMigrator) GetImage() string {
	if component.Image == nil {
		return "jmonsinjon/notary-db-migrator:v0.6.1"
	}

	return *component.Image
}

func (component *PortalComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/harbor-portal:v1.10.0"
	}

	return *component.Image
}

func (component *RegistryComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.10.0"
	}

	return *component.Image
}

func (component *RegistryControllerComponent) GetImage() string {
	if component.Image == nil {
		return "goharbor/harbor-registryctl:v1.10.0"
	}

	return *component.Image
}

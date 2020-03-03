package v1alpha2

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/goharbor/harbor-operator/pkg/images"
	"github.com/pkg/errors"
)

func (component *CoreSpec) GetImage() (string, error) {
	return fmt.Sprintf("goharbor/harbor-core:v%s", component.Version), nil
}

func (component *RegistryControllerSpec) GetImage() (string, error) {
	return fmt.Sprintf("goharbor/harbor-registryctl:v%s", component.Version), nil
}

func (component *NotaryServerSpec) GetDBMigratorImage() (string, error) {
	return fmt.Sprintf("jmonsinjon/notary-db-migrator:v%s", component.Version), nil
}

func (component *NotarySignerSpec) GetDBMigratorImage() (string, error) {
	return fmt.Sprintf("jmonsinjon/notary-db-migrator:v%s", component.Version), nil
}

func (component *JobServiceSpec) GetImage() (string, error) {
	return fmt.Sprintf("goharbor/harbor-jobservice:v%s", component.Version), nil
}

func (component *PortalSpec) GetImage() (string, error) {
	return fmt.Sprintf("goharbor/harbor-portal:v%s", component.Version), nil
}

var (
	ChartMuseumImages images.Store
)

func init() {
	{
		version := ">= 1.10.0"

		versionRange := semver.MustParseRange(version)

		err := ChartMuseumImages.AddImage(versionRange, "goharbor/chartmuseum-photon:v0.9.0-v{{.version}}")
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s image for version %s", "ChartMuseum", version))
		}
	}
}

func (component *ChartMuseumSpec) GetImage() (string, error) {
	return ChartMuseumImages.GetTag(semver.MustParse(component.Version))
}

var (
	ClairImages images.Store
)

func init() {
	{
		version := ">= 1.10.0"

		versionRange := semver.MustParseRange(version)

		err := ClairImages.AddImage(versionRange, "goharbor/clair-photon:v2.1.1-v{{.version}}")
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s image for version %s", "Clair", version))
		}
	}
}

func (component *ClairSpec) GetImage() (string, error) {
	return ClairImages.GetTag(semver.MustParse(component.Version))
}

var (
	ClairAdapterImages images.Store
)

func init() {
	{
		version := ">= 1.10.0"

		versionRange := semver.MustParseRange(version)

		// Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
		err := ClairAdapterImages.AddImage(versionRange, "holyhope/clair-adapter-with-config:v1.10.0")
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s image for version %s", "ClairAdapter", version))
		}
	}
}

func (component *ClairSpec) GetAdapterImage() (string, error) {
	return ClairAdapterImages.GetTag(semver.MustParse(component.Version))
}

var (
	NotaryServerImages images.Store
)

func init() {
	{
		version := ">= 1.10.0"

		versionRange := semver.MustParseRange(version)

		// Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
		err := NotaryServerImages.AddImage(versionRange, "goharbor/notary-server-photon:v0.6.1-v{{.version}}")
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s image for version %s", "NotaryServer", version))
		}
	}
}

func (component *NotaryServerSpec) GetImage() (string, error) {
	return NotaryServerImages.GetTag(semver.MustParse(component.Version))
}

var (
	NotarySignerImages images.Store
)

func init() {
	{
		version := ">= 1.10.0"

		versionRange := semver.MustParseRange(version)

		// Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
		err := NotarySignerImages.AddImage(versionRange, "goharbor/notary-signer-photon:v0.6.1-v{{.version}}")
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s image for version %s", "NotarySigner", version))
		}
	}
}

func (component *NotarySignerSpec) GetImage() (string, error) {
	return NotarySignerImages.GetTag(semver.MustParse(component.Version))
}

var (
	RegistryImages images.Store
)

func init() {
	{
		version := ">= 1.10.0"

		versionRange := semver.MustParseRange(version)

		// Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
		err := RegistryImages.AddImage(versionRange, "goharbor/registry-photon:v2.7.1-patch-2819-2553-v{{.version}}")
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s image for version %s", "Registry", version))
		}
	}
}

func (component *RegistrySpec) GetImage() (string, error) {
	return RegistryImages.GetTag(semver.MustParse(component.Version))
}

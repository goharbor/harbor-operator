package v1alpha1

import (
	"sync"

	"github.com/pkg/errors"
)

type Images struct {
	Core         string
	Registry     string
	RegistryCtl  string
	Portal       string
	JobService   string
	ChartMuseum  string
	Clair        string
	ClairAdapter string
	Notary       string
}

var (
	compatibilities sync.Map
)

func RegisterVersion(version string, images *Images) error {
	_, exists := compatibilities.LoadOrStore(version, images)
	if exists {
		return errors.Errorf("version %s already registered", version)
	}

	return nil
}

func GetImages(version string) (*Images, error) {
	images, ok := compatibilities.Load(version)
	if !ok {
		return nil, errors.Errorf("version %s not registered", version)
	}

	return images.(*Images), nil
}

func RegisterDefaultVersion() {
	_ = RegisterVersion("1.9.0", &Images{
		Core:         "goharbor/harbor-core:v1.9.0",
		Registry:     "goharbor/registry-photon:v2.7.1-patch-2819-v1.9.0",
		RegistryCtl:  "goharbor/harbor-registryctl:v1.9.0",
		Portal:       "goharbor/harbor-portal:v1.9.0",
		JobService:   "goharbor/harbor-jobservice:v1.9.0",
		ChartMuseum:  "goharbor/chartmuseum-photon:v0.9.0-v1.9.0",
		Clair:        "goharbor/clair-photon:v2.0.9-v1.9.0",
		ClairAdapter: "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	})
	_ = RegisterVersion("1.9.1", &Images{
		Core:         "goharbor/harbor-core:v1.9.1",
		Registry:     "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.9.1",
		RegistryCtl:  "goharbor/harbor-registryctl:v1.9.1",
		Portal:       "goharbor/harbor-portal:v1.9.1",
		JobService:   "goharbor/harbor-jobservice:v1.9.1",
		ChartMuseum:  "goharbor/chartmuseum-photon:v0.9.0-v1.9.1",
		Clair:        "goharbor/clair-photon:v2.0.9-v1.9.1",
		ClairAdapter: "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	})
	_ = RegisterVersion("1.9.2", &Images{
		Core:         "goharbor/harbor-core:v1.9.2",
		Registry:     "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.9.2",
		RegistryCtl:  "goharbor/harbor-registryctl:v1.9.2",
		Portal:       "goharbor/harbor-portal:v1.9.2",
		JobService:   "goharbor/harbor-jobservice:v1.9.2",
		ChartMuseum:  "goharbor/chartmuseum-photon:v0.9.0-v1.9.2",
		Clair:        "goharbor/clair-photon:v2.0.9-v1.9.2",
		ClairAdapter: "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	})
	_ = RegisterVersion("1.9.3", &Images{
		Core:         "goharbor/harbor-core:v1.9.3",
		Registry:     "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.9.3",
		RegistryCtl:  "goharbor/harbor-registryctl:v1.9.3",
		Portal:       "goharbor/harbor-portal:v1.9.3",
		JobService:   "goharbor/harbor-jobservice:v1.9.3",
		ChartMuseum:  "goharbor/chartmuseum-photon:v0.9.0-v1.9.3",
		Clair:        "goharbor/clair-photon:v2.1.0-v1.9.3",
		ClairAdapter: "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	})
	_ = RegisterVersion("1.9.4", &Images{
		Core:         "goharbor/harbor-core:v1.9.4",
		Registry:     "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.9.4",
		RegistryCtl:  "goharbor/harbor-registryctl:v1.9.4",
		Portal:       "goharbor/harbor-portal:v1.9.4",
		JobService:   "goharbor/harbor-jobservice:v1.9.4",
		ChartMuseum:  "goharbor/chartmuseum-photon:v0.9.0-v1.9.4",
		Clair:        "goharbor/clair-photon:v2.1.0-v1.9.4",
		ClairAdapter: "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	})
	_ = RegisterVersion("1.10.0", &Images{
		Core:         "goharbor/harbor-core:v1.10.0",
		Registry:     "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.10.0",
		RegistryCtl:  "goharbor/harbor-registryctl:v1.10.0",
		Portal:       "goharbor/harbor-portal:v1.10.0",
		JobService:   "goharbor/harbor-jobservice:v1.10.0",
		ChartMuseum:  "goharbor/chartmuseum-photon:v0.9.0-v1.10.0",
		Clair:        "goharbor/clair-photon:v2.1.1-v1.10.0",
		ClairAdapter: "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
	})
}

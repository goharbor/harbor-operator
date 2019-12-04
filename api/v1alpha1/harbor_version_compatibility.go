package v1alpha1

import (
	"sync"

	"github.com/pkg/errors"
)

type Images struct {
	Core        string
	Registry    string
	RegistryCtl string
	Portal      string
	JobService  string
	ChartMuseum string
	Clair       string
	Notary      string
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
		Core:        "goharbor/harbor-core:v1.9.0",
		Registry:    "goharbor/registry-photon:v2.7.1-patch-2819-v1.9.0",
		RegistryCtl: "goharbor/harbor-registryctl:v1.9.0",
		Portal:      "goharbor/harbor-portal:v1.9.0",
		JobService:  "goharbor/harbor-jobservice:v1.9.0",
		ChartMuseum: "goharbor/chartmuseum-photon:v0.9.0-v1.9.0",
		Clair:       "goharbor/clair-photon:v2.0.9-v1.9.0",
	})
	_ = RegisterVersion("1.9.1", &Images{
		Core:        "goharbor/harbor-core:v1.9.1",
		Registry:    "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.9.1",
		RegistryCtl: "goharbor/harbor-registryctl:v1.9.1",
		Portal:      "goharbor/harbor-portal:v1.9.1",
		JobService:  "goharbor/harbor-jobservice:v1.9.1",
		ChartMuseum: "goharbor/chartmuseum-photon:v0.9.0-v1.9.1",
		Clair:       "goharbor/clair-photon:v2.0.9-v1.9.1",
	})
}

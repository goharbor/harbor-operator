package v1alpha1

import (
	"sync"

	"github.com/pkg/errors"
)

type Images struct {
	Core             string
	Registry         string
	RegistryCtl      string
	Portal           string
	JobService       string
	ChartMuseum      string
	Clair            string
	ClairAdapter     string
	NotarySigner     string
	NotaryServer     string
	NotaryDBMigrator string
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
	_ = RegisterVersion("1.10.0", &Images{
		Core:             "goharbor/harbor-core:v1.10.0",
		Registry:         "goharbor/registry-photon:v2.7.1-patch-2819-2553-v1.10.0",
		RegistryCtl:      "goharbor/harbor-registryctl:v1.10.0",
		Portal:           "goharbor/harbor-portal:v1.10.0",
		JobService:       "goharbor/harbor-jobservice:v1.10.0",
		ChartMuseum:      "goharbor/chartmuseum-photon:v0.9.0-v1.10.0",
		Clair:            "goharbor/clair-photon:v2.1.1-v1.10.0",
		ClairAdapter:     "holyhope/clair-adapter-with-config:v1.10.0", // Use "goharbor/clair-adapter-photon:v1.0.1-v1.10.0" when possible
		NotarySigner:     "goharbor/notary-signer-photon:v0.6.1-v1.10.0",
		NotaryServer:     "goharbor/notary-server-photon:v0.6.1-v1.10.0",
		NotaryDBMigrator: "jmonsinjon/notary-db-migrator:v0.6.1",
	})
}

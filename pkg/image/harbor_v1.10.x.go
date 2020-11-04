// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"fmt"
	"strings"
)

const (
	defaultVersion   = "v1.10.4"
	defaultNS        = "goharbor"
	portalRepo       = "harbor-portal"
	coreRepo         = "harbor-core"
	registryRepo     = "registry-photon"
	jobserviceRepo   = "harbor-jobservice"
	chartmuseumRepo  = "chartmuseum-photon"
	registryCtlRepo  = "harbor-registryctl"
	notaryServerRepo = "notary-server-photon"
	notarySignerRepo = "notary-signer-photon"
	clairRepo        = "clair-photon"
	clairAdapterRepo = "clair-adapter-photon"
	// A special image for handling notary data migration
	migratorRepo = "jmonsinjon/notary-db-migrator:v0.6.1"
)

// harborVM1m10pxImageLocator supports version > 1.10.1
type harborVM1m10pxImageLocator struct {
	// Version of Harbor
	HarborVersion string
}

func (dil *harborVM1m10pxImageLocator) version() string {
	if len(dil.HarborVersion) == 0 {
		return defaultVersion
	}

	if !strings.HasPrefix(dil.HarborVersion, "v") {
		return fmt.Sprintf("v%s", dil.HarborVersion)
	}

	return dil.HarborVersion
}

func (dil *harborVM1m10pxImageLocator) imagePath(component string) string {
	return fmt.Sprintf("%s/%s:%s", defaultNS, component, dil.version())
}

func (dil *harborVM1m10pxImageLocator) CoreImage() string {
	return dil.imagePath(coreRepo)
}

func (dil *harborVM1m10pxImageLocator) ChartMuseumImage() string {
	return dil.imagePath(chartmuseumRepo)
}

func (dil *harborVM1m10pxImageLocator) ClairImage() string {
	return dil.imagePath(clairRepo)
}

func (dil *harborVM1m10pxImageLocator) ClairAdapterImage() string {
	return dil.imagePath(clairAdapterRepo)
}

func (dil *harborVM1m10pxImageLocator) JobServiceImage() string {
	return dil.imagePath(jobserviceRepo)
}

func (dil *harborVM1m10pxImageLocator) NotaryServerImage() string {
	return dil.imagePath(notaryServerRepo)
}

func (dil *harborVM1m10pxImageLocator) NotarySingerImage() string {
	return dil.imagePath(notarySignerRepo)
}

func (dil *harborVM1m10pxImageLocator) NotaryDBMigratorImage() string {
	return migratorRepo
}

func (dil *harborVM1m10pxImageLocator) PortalImage() string {
	return dil.imagePath(portalRepo)
}

func (dil *harborVM1m10pxImageLocator) RegistryImage() string {
	return dil.imagePath(registryRepo)
}

func (dil *harborVM1m10pxImageLocator) RegistryControllerImage() string {
	return dil.imagePath(registryCtlRepo)
}

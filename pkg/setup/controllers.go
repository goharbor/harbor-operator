package setup

import (
	"context"

	"github.com/ovh/configstore"

	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
	"github.com/goharbor/harbor-operator/controllers/goharbor/jobservice"
	notaryserver "github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	notarysigner "github.com/goharbor/harbor-operator/controllers/goharbor/notarysigner"
	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registryctl"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
)

//go:generate stringer -type=ControllerUID -linecomment
type ControllerUID int

const (
	Harbor       ControllerUID = iota // harbor
	Core                              // core
	JobService                        // jobservice
	Registry                          // registry
	NotaryServer                      // notary-server
	NotarySigner                      // notary-signer
	RegistryCtl                       // registryctl
	Portal                            // portal
	ChartMuseum                       // chartmuseum
)

var controllers = map[ControllerUID]func(context.Context, string, string, *configstore.Store) (commonCtrl.Reconciler, error){
	Core:         core.New,
	Harbor:       harbor.New,
	JobService:   jobservice.New,
	Registry:     registry.New,
	NotaryServer: notaryserver.New,
	NotarySigner: notarysigner.New,
	RegistryCtl:  registryctl.New,
	Portal:       portal.New,
	ChartMuseum:  chartmuseum.New,
}

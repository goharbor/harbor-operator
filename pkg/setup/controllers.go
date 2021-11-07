package setup

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/configuration"
	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/exporter"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborserverconfiguration"
	"github.com/goharbor/harbor-operator/controllers/goharbor/jobservice"
	"github.com/goharbor/harbor-operator/controllers/goharbor/namespace"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notarysigner"
	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/controllers/goharbor/pullsecretbinding"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/controllers/goharbor/trivy"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	ControllerDisabledSuffixConfigKey = "controller-disabled"
)

var controllersBuilder = map[controllers.Controller]func(context.Context, *configstore.Store) (commonCtrl.Reconciler, error){
	controllers.Core:          core.New,
	controllers.Exporter:      exporter.New,
	controllers.Harbor:        harbor.New,
	controllers.JobService:    jobservice.New,
	controllers.Registry:      registry.New,
	controllers.NotaryServer:  notaryserver.New,
	controllers.NotarySigner:  notarysigner.New,
	controllers.Portal:        portal.New,
	controllers.ChartMuseum:   chartmuseum.New,
	controllers.Trivy:         trivy.New,
	controllers.HarborCluster: harborcluster.New,
	// old configmap controller is planned to be removed at v1.3,
	// the controller converts the cm to configuration cr.
	controllers.HarborConfigurationCm:     configuration.NewWithCm,
	controllers.HarborConfiguration:       configuration.New,
	controllers.HarborServerConfiguration: harborserverconfiguration.New,
	controllers.PullSecretBinding:         pullsecretbinding.New,
	controllers.Namespace:                 namespace.New,
}

type ControllerFactory func(context.Context, string, string, *configstore.Store) (commonCtrl.Reconciler, error)

type Controller interface {
	WithManager(context.Context, manager.Manager) (commonCtrl.Reconciler, error)
	IsEnabled(context.Context) (bool, error)
}

type controller struct {
	Name controllers.Controller
	New  func(context.Context, *configstore.Store) (commonCtrl.Reconciler, error)
}

func NewController(name controllers.Controller, factory func(context.Context, *configstore.Store) (commonCtrl.Reconciler, error)) Controller {
	return &controller{Name: name, New: factory}
}

func (c *controller) GetConfig(ctx context.Context) (*configstore.Store, error) {
	configStore := config.NewConfigWithDefaults()

	configStore.RegisterProvider("common", func() (configstore.ItemList, error) {
		itemList, err := configstore.DefaultStore.GetItemList()
		if err != nil {
			return configstore.ItemList{}, errors.Wrap(err, "item list from default")
		}

		return *itemList, nil
	})

	configStore.Env(c.Name.String())

	configDirectory, err := config.GetString(configstore.DefaultStore, config.CtrlConfigDirectoryKey, config.DefaultConfigDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "config directory")
	}

	configPath := path.Join(configDirectory, fmt.Sprintf("%s-ctrl.yaml", c.Name.String()))

	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "invalid config file")
		}

		logger.Get(ctx).Error(err, "invalid config file")
	} else {
		configStore.FileRefresh(configPath)
	}

	return configStore, nil
}

func (c *controller) WithManager(ctx context.Context, mgr manager.Manager) (commonCtrl.Reconciler, error) {
	configStore, err := c.GetConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get configuration")
	}

	controller, err := c.New(ctx, configStore)
	if err != nil {
		return controller, errors.Wrap(err, "create")
	}

	err = controller.SetupWithManager(ctx, mgr)

	return controller, errors.Wrap(err, "setup")
}

func (c *controller) IsEnabled(ctx context.Context) (bool, error) {
	configKey := fmt.Sprintf("%s-%s", c.Name, ControllerDisabledSuffixConfigKey)

	disabled, err := configstore.GetItemValueBool(configKey)
	if err == nil {
		return !disabled, nil
	}

	if config.IsNotFound(err, configKey) {
		return true, nil
	}

	return false, err
}

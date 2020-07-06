package setup

import (
	"context"
	"fmt"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
)

const (
	ControllerDisabledSuffixConfigKey = "controller-disabled"
)

type ControllerFactory func(context.Context, string, string, *configstore.Store) (commonCtrl.Reconciler, error)

type Controller interface {
	WithManager(context.Context, manager.Manager) error
	IsEnabled(context.Context) (bool, error)
}

type controller struct {
	Name ControllerUID
	New  func(context.Context, string, *configstore.Store) (commonCtrl.Reconciler, error)
}

func (c *controller) GetConfig(ctx context.Context) (*configstore.Store, error) {
	configStore := config.NewConfigWithDefaults()
	configStore.Env(c.Name.String())

	return configStore, nil
}

func (c *controller) WithManager(ctx context.Context, mgr manager.Manager) error {
	configStore, err := c.GetConfig(ctx)
	if err != nil {
		return errors.Wrap(err, "get configuration")
	}

	controller, err := c.New(ctx, c.Name.String(), configStore)
	if err != nil {
		return errors.Wrap(err, "create")
	}

	err = controller.SetupWithManager(ctx, mgr)

	return errors.Wrap(err, "setup")
}

func (c *controller) IsEnabled(ctx context.Context) (bool, error) {
	disabled, err := configstore.GetItemValueBool(fmt.Sprintf("%s-%s", c.Name, ControllerDisabledSuffixConfigKey))
	if err == nil {
		return !disabled, nil
	}

	if _, ok := err.(configstore.ErrItemNotFound); ok {
		return true, nil
	}

	return false, err
}

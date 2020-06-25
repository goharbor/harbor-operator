package setup

import (
	"context"
	"fmt"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
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
	New  func(context.Context, string, string, *configstore.Store) (commonCtrl.Reconciler, error)
}

func (c *controller) WithManager(ctx context.Context, mgr manager.Manager) error {
	controller, err := c.New(ctx, c.Name.String(), application.GetVersion(ctx), config.NewConfigWithDefaults())
	if err != nil {
		return errors.Wrap(err, "create")
	}

	err = controller.SetupWithManager(ctx, mgr)

	return errors.Wrap(err, "setup")
}

func (c *controller) IsEnabled(ctx context.Context) (bool, error) {
	ok, err := configstore.GetItemValueBool(fmt.Sprintf("%s-%s", c.Name, ControllerDisabledSuffixConfigKey))
	if err == nil {
		return ok, nil
	}

	if _, ok := err.(configstore.ErrItemNotFound); ok {
		return true, nil
	}

	return false, err
}

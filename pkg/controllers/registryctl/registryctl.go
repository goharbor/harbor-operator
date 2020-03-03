package registryctl

import (
	"context"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor-operator/controllers/registryctl"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name         = "registryctl"
	ConfigPrefix = Name + "-controller"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	config, err := config.GetConfig(ConfigPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get configuration")
	}

	return registryctl.New(ctx, Name, version, config)
}

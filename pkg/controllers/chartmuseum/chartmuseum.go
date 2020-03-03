package chartmuseum

import (
	"context"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor-operator/controllers/chartmuseum"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name         = "chartmuseum"
	ConfigPrefix = Name + "-controller"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	config, err := config.GetConfig(ConfigPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get configuration")
	}

	return chartmuseum.New(ctx, Name, version, config)
}

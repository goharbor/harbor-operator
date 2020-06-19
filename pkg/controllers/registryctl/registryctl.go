package registryctl

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/registryctl"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "registryctl"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return registryctl.New(ctx, Name, version, config.NewConfigWithDefaults())
}

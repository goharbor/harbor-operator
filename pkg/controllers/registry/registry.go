package registry

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "registry"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return registry.New(ctx, Name, version, config.NewConfigWithDefaults())
}

package portal

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "portal"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return portal.New(ctx, Name, version, config.NewConfigWithDefaults())
}

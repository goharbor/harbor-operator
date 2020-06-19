package chartmuseum

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "chartmuseum"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return chartmuseum.New(ctx, Name, version, config.NewConfigWithDefaults())
}

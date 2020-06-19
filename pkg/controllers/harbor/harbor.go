package harbor

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "harbor"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return harbor.New(ctx, Name, version, config.NewConfigWithDefaults())
}

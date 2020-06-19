package core

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "core"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return core.New(ctx, Name, version, config.NewConfigWithDefaults())
}

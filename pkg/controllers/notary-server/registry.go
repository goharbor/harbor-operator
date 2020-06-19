package notaryserver

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "notary-server"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return notaryserver.New(ctx, Name, version, config.NewConfigWithDefaults())
}

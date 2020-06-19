package jobservice

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers/goharbor/jobservice"
	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
)

const (
	Name = "jobservice"
)

func New(ctx context.Context, version string) (controllers.Controller, error) {
	return jobservice.New(ctx, Name, version, config.NewConfigWithDefaults())
}

package controller

import (
	"context"

	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (c *Controller) NewContext(req ctrl.Request) context.Context {
	ctx := context.TODO()
	application.SetName(&ctx, c.GetName())
	application.SetVersion(&ctx, c.GetVersion())
	application.SetGitCommit(&ctx, c.GetGitCommit())
	ctx = sgraph.WithGraph(ctx)

	logger.Set(&ctx, c.Log.WithValues("request", req))

	return ctx
}

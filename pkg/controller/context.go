package controller

import (
	"context"

	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (c *Controller) PopulateContext(ctx context.Context, req ctrl.Request) context.Context {
	application.SetName(&ctx, c.GetName())
	application.SetVersion(&ctx, c.GetVersion())
	application.SetGitCommit(&ctx, c.GetGitCommit())
	sgraph.SetGraph(&ctx, graph.NewResourceManager())

	logger.Set(&ctx, c.Log.WithValues("request", req))

	return ctx
}

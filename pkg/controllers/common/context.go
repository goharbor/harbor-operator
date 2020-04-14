package common

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	sgraph "github.com/goharbor/harbor-operator/pkg/controllers/common/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

func (c *Controller) PopulateContext(ctx context.Context, req ctrl.Request) context.Context {
	ctx = sgraph.WithGraph(ctx)
	application.SetName(&ctx, c.GetName())
	application.SetVersion(&ctx, c.GetVersion())

	l := logger.Get(ctx).WithValues(
		"Request.Namespace", req.Namespace,
		"Request.Name", req.Name,
	)
	logger.Set(&ctx, l)

	return ctx
}

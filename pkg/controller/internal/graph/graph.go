package graph

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/graph"
)

var graphKey = "graph"

func SetGraph(ctx *context.Context, mgr graph.Manager) {
	*ctx = context.WithValue(*ctx, &graphKey, mgr)
}

func Get(ctx context.Context) graph.Manager {
	g := ctx.Value(&graphKey)
	if g == nil {
		return nil
	}

	return g.(graph.Manager)
}

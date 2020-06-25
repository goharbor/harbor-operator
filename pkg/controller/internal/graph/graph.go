package graph

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/graph"
)

var graphKey = "graph"

func WithGraph(ctx context.Context) context.Context {
	g := graph.NewResourceManager()
	return context.WithValue(ctx, &graphKey, g)
}

func Get(ctx context.Context) graph.Manager {
	g := ctx.Value(&graphKey)
	if g == nil {
		return nil
	}

	return g.(graph.Manager)
}

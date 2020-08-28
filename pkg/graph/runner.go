package graph

import (
	"context"
	"sync"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/sync/errgroup"
)

func (rm *resourceManager) Run(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "walkGraph", opentracing.Tags{
		"Nodes.count": len(rm.resources),
	})
	defer span.Finish()

	g := errgroup.Group{}
	l := logger.Get(ctx)

	for _, no := range rm.getGraph(ctx) {
		no := no

		g.Go(func() (err error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "executeNode", opentracing.Tags{
				"Node": no,
			})
			defer span.Finish()

			defer func() {
				err := no.Terminates(err)
				if err != nil {
					l.Error(err, "failed to terminate node when running graph")
				}
			}()

			err = no.Wait(ctx)
			if err != nil {
				return err
			}

			err = no.fn(ctx, no.resource)

			return err
		})
	}

	return g.Wait()
}

func (rm *resourceManager) getGraph(ctx context.Context) []*node {
	span, _ := opentracing.StartSpanFromContext(ctx, "getGraph")
	defer span.Finish()

	rm.lock.Lock()
	defer rm.lock.Unlock()

	graph := make(map[Resource]*node, len(rm.resources))
	result := make([]*node, len(rm.resources))

	i := 0

	for resource, blockers := range rm.resources {
		blockerCount := len(blockers)

		node := &node{
			resource: resource,
			fn:       rm.functions[resource],

			parent:      make(chan error, blockerCount),
			parentLock:  &sync.Mutex{},
			parentCount: blockerCount,

			children:     []chan<- error{},
			childrenLock: []*sync.Mutex{},
		}
		graph[resource] = node
		result[i] = node

		i++

		blockers := blockers

		defer func() {
			for _, blocker := range blockers {
				graph[blocker].AddChild(node)
			}
		}()
	}

	return result
}

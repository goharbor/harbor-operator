package graph

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (rm *resourceManager) Run(ctx context.Context, runner func(context.Context, Resource) error) error {
	g := errgroup.Group{}

	for _, no := range rm.getGraph(ctx) {
		no := no

		g.Go(func() error {
			var err error

			defer func() {
				no.Terminates(err)
			}()

			err = no.Wait()
			if err != nil {
				return err
			}

			err = runner(ctx, no.resource)

			return err
		})
	}

	return g.Wait()
}

func (rm *resourceManager) getGraph(ctx context.Context) []*node {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	graph := make(map[Resource]*node, len(rm.resources))
	result := make([]*node, len(rm.resources))

	i := 0
	for resource, blockers := range rm.resources {
		blockerCount := len(blockers)

		c := make(chan error, blockerCount)

		node := &node{
			resource: resource,

			parent:      c,
			parentCount: blockerCount,
			children:    []chan<- error{},
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

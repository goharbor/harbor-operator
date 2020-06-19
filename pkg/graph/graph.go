package graph

import (
	"context"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type resourceManager struct {
	resources map[Resource][]Resource
	functions map[Resource]RunFunc

	lock sync.Mutex
}

func NewResourceManager() Manager {
	return &resourceManager{
		resources: map[Resource][]Resource{},
		functions: map[Resource]RunFunc{},
	}
}

func (rm *resourceManager) AddResource(ctx context.Context, resource Resource, blockers []Resource, run RunFunc) error {
	if resource == nil {
		return nil
	}

	if run == nil {
		return errors.Errorf("unsupported RunFunc value %v", run)
	}

	span, _ := opentracing.StartSpanFromContext(ctx, "addResource", opentracing.Tags{
		"Resource": resource,
	})
	defer span.Finish()

	rm.lock.Lock()
	defer rm.lock.Unlock()

	_, ok := rm.resources[resource]
	if ok {
		return errors.Errorf("resource %+v already added", resource)
	}

	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}

		_, ok := rm.resources[blocker]
		if !ok {
			return errors.Errorf("unknown blocker %+v", blocker)
		}
	}

	rm.resources[resource] = blockers
	rm.functions[resource] = run

	return nil
}

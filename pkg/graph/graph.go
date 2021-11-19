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

	lock sync.RWMutex
}

func NewResourceManager() Manager {
	return &resourceManager{
		resources: map[Resource][]Resource{},
		functions: map[Resource]RunFunc{},
	}
}

func (rm *resourceManager) GetAllResources(ctx context.Context) []Resource {
	resources := []Resource{}

	rm.lock.RLock()
	defer rm.lock.RUnlock()

	for resource := range rm.resources {
		resources = append(resources, resource)
	}

	return resources
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

	nonNilBlockers := []Resource{}

	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}

		nonNilBlockers = append(nonNilBlockers, blocker)

		_, ok := rm.resources[blocker]
		if !ok {
			return errors.Errorf("unknown blocker %+v", blocker)
		}
	}

	rm.lock.Lock()
	defer rm.lock.Unlock()

	_, ok := rm.resources[resource]
	if ok {
		return errors.Errorf("resource %+v already added", resource)
	}

	rm.resources[resource] = nonNilBlockers
	rm.functions[resource] = run

	return nil
}

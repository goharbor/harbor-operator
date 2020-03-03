package graph

import (
	"sync"

	"github.com/pkg/errors"
)

type resourceManager struct {
	resources map[Resource][]Resource

	lock sync.Mutex

	graph []*node
}

func NewResourceManager() *resourceManager {
	return &resourceManager{
		resources: map[Resource][]Resource{},
	}
}

func (rm *resourceManager) AddResource(resource Resource, blockers []Resource) error {
	if resource == nil {
		return nil
	}

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

	return nil
}

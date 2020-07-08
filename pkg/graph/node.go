package graph

import (
	"context"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type node struct {
	resource Resource
	fn       RunFunc

	parent      chan error
	parentLock  *sync.Mutex
	parentCount int

	children     []chan<- error
	childrenLock []*sync.Mutex
}

var errParentChannelClosed = errors.New("parent channel closed")

func (no *node) Wait(ctx context.Context) error {
	defer close(no.parent)

	if no.parentCount == 0 {
		return nil
	}

	span, _ := opentracing.StartSpanFromContext(ctx, "waitNode")
	defer span.Finish()

	received := 0

	for parentErr := range no.parent {
		if parentErr != nil {
			// a parent failed
			return parentErr
		}

		received++

		if received >= no.parentCount {
			return nil
		}
	}

	if received < no.parentCount {
		// parents closed the channel
		return errParentChannelClosed
	}

	return nil
}

func (no *node) Terminates(err error) (result error) {
	for _, c := range no.children {
		c := c

		go func() {
			no.parentLock.Lock()
			defer no.parentLock.Unlock()

			defer func() {
				// recover from panic caused by writing to a closed channel
				if r := recover(); r != nil {
					result = errors.Errorf("%v", r)
				}
			}()

			c <- err
		}()
	}

	return result
}

func (no *node) AddChild(child *node) {
	if no == nil {
		// Parent is nil -> not a blocker
		return
	}

	no.parentLock.Lock()
	defer no.parentLock.Unlock()

	no.children = append(no.children, child.parent)
	no.childrenLock = append(no.childrenLock, child.parentLock)
}

package graph

import "github.com/pkg/errors"

type node struct {
	resource Resource

	parent      chan error
	parentCount int

	children []chan<- error
}

func (no *node) Wait() error {
	defer close(no.parent)

	if no.parentCount == 0 {
		return nil
	}

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
		return errors.New("parent channel closed")
	}

	return nil
}

func (no *node) Terminates(err error) {
	for _, c := range no.children {
		c := c

		go func() {
			c <- err
		}()
	}
}

func (no *node) AddChild(child *node) {
	no.children = append(no.children, child.parent)
}

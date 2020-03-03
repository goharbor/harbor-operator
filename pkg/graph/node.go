package graph

type node struct {
	resource Resource

	parent      chan error
	parentCount int

	children []chan<- error
}

func (no *node) Wait() error {
	defer close(no.parent)

	for received := 0; received < no.parentCount; received++ {
		select {
		case err, ok := <-no.parent:
			if !ok {
				// parents succeeded
				return nil
			}

			if err != nil {
				// a parent failed
				return err
			}
		}
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

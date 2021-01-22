package resources

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

type Mutable func(context.Context, runtime.Object) error

func (m *Mutable) AppendMutation(mutate Mutable) {
	old := *m
	*m = func(ctx context.Context, resource runtime.Object) error {
		if err := old(ctx, resource); err != nil {
			return err
		}

		return mutate(ctx, resource)
	}
}

func (m *Mutable) PrependMutation(mutate Mutable) {
	old := *m
	*m = func(ctx context.Context, resource runtime.Object) error {
		if err := mutate(ctx, resource); err != nil {
			return err
		}

		return old(ctx, resource)
	}
}

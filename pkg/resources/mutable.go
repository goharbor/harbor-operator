package resources

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Mutable func(context.Context, runtime.Object, runtime.Object) controllerutil.MutateFn

func (m *Mutable) AppendMutation(mutate Mutable) {
	old := *m
	*m = func(ctx context.Context, resource runtime.Object, result runtime.Object) controllerutil.MutateFn {
		mutate := mutate(ctx, resource, result)
		old := old(ctx, resource, result)

		return func() error {
			err := old()
			if err != nil {
				return err
			}

			return mutate()
		}
	}
}

func (m *Mutable) PrependMutation(mutate Mutable) {
	old := *m
	*m = func(ctx context.Context, resource runtime.Object, result runtime.Object) controllerutil.MutateFn {
		mutate := mutate(ctx, resource, result)
		old := old(ctx, resource, result)

		return func() error {
			err := mutate()
			if err != nil {
				return err
			}

			return old()
		}
	}
}

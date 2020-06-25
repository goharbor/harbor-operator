package mutation

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NoOp(_ context.Context, _, _ runtime.Object) controllerutil.MutateFn {
	return func() error {
		return nil
	}
}

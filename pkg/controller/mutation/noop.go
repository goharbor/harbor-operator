package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ resources.Mutable = NoOp

func NoOp(_ context.Context, _ runtime.Object) error {
	return nil
}

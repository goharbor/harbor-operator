package resources

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource interface {
	client.Object
}

type Checkable func(context.Context, client.Object) (bool, error)

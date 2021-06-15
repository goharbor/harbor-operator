package statuscheck

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func True(context.Context, client.Object) (bool, error) {
	return true, nil
}

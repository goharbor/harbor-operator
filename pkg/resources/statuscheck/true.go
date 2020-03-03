package statuscheck

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

func True(context.Context, runtime.Object) (bool, error) {
	return true, nil
}

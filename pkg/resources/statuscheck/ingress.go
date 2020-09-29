package statuscheck

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

func IngressCheck(ctx context.Context, object runtime.Object) (bool, error) {
	// Cannot use status.LoadBalancer since all ingressControllers does not update this field
	return True(ctx, object)
}

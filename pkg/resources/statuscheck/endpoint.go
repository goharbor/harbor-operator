package statuscheck

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func EndpointCheck(ctx context.Context, object runtime.Object, portNames ...string) (bool, error) {
	endpoints := object.(*corev1.Endpoints)

	ports := make(map[string]bool, len(portNames))
	for _, port := range portNames {
		ports[port] = false
	}

	for _, subset := range endpoints.Subsets {
		if len(subset.NotReadyAddresses) > 0 {
			return false, nil
		}

		for _, port := range subset.Ports {
			if v, ok := ports[port.Name]; !v && ok {
				ports[port.Name] = true
			}
		}
	}

	for _, found := range ports {
		if !found {
			return false, nil
		}
	}

	return true, nil
}

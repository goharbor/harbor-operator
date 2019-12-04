package portal

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

func (*Portal) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	return []*corev1.ConfigMap{}
}

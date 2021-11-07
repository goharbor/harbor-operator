package namespace

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &corev1.Namespace{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	return nil
}

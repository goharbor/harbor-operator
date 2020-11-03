package harborcluster

import (
	"context"
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.HarborCluster{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	return nil
}

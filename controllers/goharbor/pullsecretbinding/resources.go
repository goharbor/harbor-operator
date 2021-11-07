package pullsecretbinding

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1.PullSecretBinding{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {

	return nil
}

package harbor

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

func (r *Reconciler) InitResources() error {
	return nil
}

func (r *Reconciler) AddResources(ctx context.Context, harbor *goharborv1alpha2.Harbor) error {
	coreIngress, err := r.GetCoreIngresse(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot get core ingress")
	}

	_, err = r.Controller.AddBasicObjectToManage(ctx, coreIngress)
	if err != nil {
		return errors.Wrapf(err, "cannot add core ingress %+v", coreIngress)
	}

	notaryIngress, err := r.GetNotaryServerIngresse(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot get notary ingress")
	}

	_, err = r.Controller.AddBasicObjectToManage(ctx, notaryIngress)
	if err != nil {
		return errors.Wrapf(err, "cannot add notary ingress %+v", notaryIngress)
	}

	return nil
}

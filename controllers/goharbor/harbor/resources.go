package harbor

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controllers/common/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.Harbor{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	harbor, ok := resource.(*goharborv1alpha2.Harbor)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	coreIngress, err := r.GetCoreIngresse(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot get core ingress")
	}

	_, err = r.Controller.AddIngressToManage(ctx, coreIngress)
	if err != nil {
		return errors.Wrapf(err, "cannot add core ingress %s", coreIngress.GetName())
	}

	notaryIngress, err := r.GetNotaryServerIngresse(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot get notary ingress")
	}

	_, err = r.Controller.AddIngressToManage(ctx, notaryIngress)
	if err != nil {
		return errors.Wrapf(err, "cannot add notary ingress %s", notaryIngress.GetName())
	}

	return nil
}

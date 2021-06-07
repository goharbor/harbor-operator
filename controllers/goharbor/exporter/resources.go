package exporter

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1.Exporter{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	exporter, ok := resource.(*goharborv1.Exporter)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	service, err := r.GetService(ctx, exporter)
	if err != nil {
		return errors.Wrap(err, "get service")
	}

	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "add service %s", service.GetName())
	}

	deployment, err := r.GetDeployment(ctx, exporter)
	if err != nil {
		return errors.Wrap(err, "get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment)
	if err != nil {
		return errors.Wrapf(err, "add deployment %s", deployment.GetName())
	}

	err = r.AddNetworkPolicies(ctx, exporter)

	return errors.Wrap(err, "add network policies")
}

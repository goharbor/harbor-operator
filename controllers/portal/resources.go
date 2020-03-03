package portal

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

func (r *Reconciler) InitResources() error {
	return nil
}

func (r *Reconciler) AddResources(ctx context.Context, portal *goharborv1alpha2.Portal) error {
	service, err := r.GetService(ctx, portal)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddBasicObjectToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %+v", service)
	}

	deployment, err := r.GetDeployment(ctx, portal)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %+v", deployment)
	}

	return nil
}

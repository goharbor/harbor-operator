package portal

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1.Portal{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	portal, ok := resource.(*goharborv1.Portal)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	service, err := r.GetService(ctx, portal)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	configMap, err := r.GetConfigMap(ctx, portal)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapRes, err := r.Controller.AddConfigMapToManage(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", configMap.GetName())
	}

	deployment, err := r.GetDeployment(ctx, portal)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, configMapRes)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	err = r.AddNetworkPolicies(ctx, portal)

	return errors.Wrap(err, "network policies")
}

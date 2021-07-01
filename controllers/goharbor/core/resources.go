package core

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1.Core{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	core, ok := resource.(*goharborv1.Core)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	service, err := r.GetService(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	configMap, err := r.GetConfigMap(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.Controller.AddConfigMapToManage(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", configMap.GetName())
	}

	secret, err := r.GetSecret(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get secret")
	}

	secretResource, err := r.Controller.AddSecretToManage(ctx, secret)
	if err != nil {
		return errors.Wrapf(err, "cannot add secret %s", secret.GetName())
	}

	deployment, err := r.GetDeployment(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, configMapResource, secretResource)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	err = r.AddNetworkPolicies(ctx, core)

	return errors.Wrap(err, "network policies")
}

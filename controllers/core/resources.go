package core

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

func (r *Reconciler) InitResources() error {
	err := r.InitConfigMaps()
	return errors.Wrap(err, "configMaps")
}

func (r *Reconciler) AddResources(ctx context.Context, core *goharborv1alpha2.Core) error {
	service, err := r.GetService(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddBasicObjectToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %+v", service)
	}

	configMap, err := r.GetConfigMap(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.Controller.AddInstantResourceToManage(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %+v", configMap)
	}

	secret, err := r.GetSecret(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get secret")
	}

	secretResource, err := r.Controller.AddInstantResourceToManage(ctx, secret)
	if err != nil {
		return errors.Wrapf(err, "cannot add secret %+v", secret)
	}

	deployment, err := r.GetDeployment(ctx, core)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, configMapResource, secretResource)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %+v", deployment)
	}

	return nil
}

package clair

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

func (r *Reconciler) InitResources() error {
	return errors.Wrap(r.InitConfigMaps(), "configMaps")
}

func (r *Reconciler) AddResources(ctx context.Context, clair *goharborv1alpha2.Clair) error {
	service, err := r.GetService(ctx, clair)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddBasicObjectToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %+v", service)
	}

	configMap, err := r.GetConfigMap(ctx, clair)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.Controller.AddInstantResourceToManage(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %+v", configMap)
	}

	deployment, err := r.GetDeployment(ctx, clair)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, configMapResource)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %+v", deployment)
	}

	return nil
}

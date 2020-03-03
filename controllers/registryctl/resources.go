package registryctl

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

func (r *Reconciler) InitResources() error {
	return errors.Wrap(r.InitConfigMaps(), "configmaps")
}

func (r *Reconciler) AddResources(ctx context.Context, registryctl *goharborv1alpha2.RegistryController) error {
	cm, err := r.GetConfigMap(ctx, registryctl)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	_, err = r.Controller.AddInstantResourceToManage(ctx, cm)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %+v", cm)
	}

	return errors.New("not yet implemented")
}

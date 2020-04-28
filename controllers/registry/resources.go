package registry

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

func (r *Reconciler) InitResources() error {
	return errors.Wrap(r.InitConfigMaps(), "configmaps")
}

func (r *Reconciler) AddResources(ctx context.Context, registry *goharborv1alpha2.Registry) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "addResources", opentracing.Tags{})
	defer span.Finish()

	service, err := r.GetService(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddBasicObjectToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	configMap, err := r.GetConfigMap(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.Controller.AddInstantResourceToManage(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", configMap.GetName())
	}

	secret, err := r.GetSecret(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get secret")
	}

	secretResource, err := r.Controller.AddInstantResourceToManage(ctx, secret)
	if err != nil {
		return errors.Wrapf(err, "cannot add secret %s", secret.GetName())
	}

	certificate, err := r.GetCertificate(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get certificate")
	}

	certificateResource, err := r.Controller.AddCertificateToManage(ctx, certificate)
	if err != nil {
		return errors.Wrapf(err, "cannot add certificate %s", certificate.GetName())
	}

	deployment, err := r.GetDeployment(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, configMapResource, secretResource, certificateResource)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	return nil
}

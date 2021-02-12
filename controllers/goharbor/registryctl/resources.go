package registryctl

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.RegistryController{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	registryctl, ok := resource.(*goharborv1alpha2.RegistryController)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	service, err := r.GetService(ctx, registryctl)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	registry, err := r.AddExternalResource(ctx, &goharborv1alpha2.Registry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryctl.Spec.RegistryRef,
			Namespace: registryctl.GetNamespace(),
		},
	})
	if err != nil {
		return errors.Wrapf(err, "cannot add registry %s", registryctl.Spec.RegistryRef)
	}

	_, err = r.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	cm, err := r.GetConfigMap(ctx, registryctl)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.AddConfigMapToManage(ctx, cm)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", cm.GetName())
	}

	deployment, err := r.GetDeployment(ctx, registryctl)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.AddDeploymentToManage(ctx, deployment, registry, configMapResource)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	err = r.AddNetworkPolicies(ctx, registryctl)

	return errors.Wrap(err, "network policies")
}

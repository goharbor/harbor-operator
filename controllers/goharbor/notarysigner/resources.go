package notarysigner

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1.NotarySigner{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	notary, ok := resource.(*goharborv1.NotarySigner)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	service, err := r.GetService(ctx, notary)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	_, err = r.Controller.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	var storageSecret graph.Resource

	if notary.Spec.Storage.Postgres.PasswordRef != "" {
		storageSecret, err = r.AddExternalTypedSecret(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notary.Spec.Storage.Postgres.PasswordRef,
				Namespace: notary.GetNamespace(),
			},
		}, harbormetav1.SecretTypePostgresql)
		if err != nil {
			return errors.Wrap(err, "cannot add migration secret")
		}
	}

	configMap, err := r.GetConfigMap(ctx, notary)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.Controller.AddConfigMapToManage(ctx, configMap, storageSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", configMap.GetName())
	}

	deployment, err := r.GetDeployment(ctx, notary)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.Controller.AddDeploymentToManage(ctx, deployment, configMapResource)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	return nil
}

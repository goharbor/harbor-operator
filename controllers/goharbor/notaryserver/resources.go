package notaryserver

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.NotaryServer{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error { // nolint:funlen
	notaryserver, ok := resource.(*goharborv1alpha2.NotaryServer)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	service, err := r.GetService(ctx, notaryserver)
	if err != nil {
		return errors.Wrap(err, "cannot get service")
	}

	var migrationSecret graph.Resource

	if notaryserver.Spec.Migration.Enabled() && notaryserver.Spec.Migration.Github != nil {
		migrationSecret, err = r.AddExternalResource(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notaryserver.Spec.Migration.Github.CredentialsRef,
				Namespace: notaryserver.GetNamespace(),
			},
		})
		if err != nil {
			return errors.Wrap(err, "cannot add migration secret")
		}
	}

	var storageSecret graph.Resource

	if notaryserver.Spec.Storage.Postgres.PasswordRef != "" {
		storageSecret, err = r.AddExternalResource(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notaryserver.Spec.Storage.Postgres.PasswordRef,
				Namespace: notaryserver.GetNamespace(),
			},
		})
		if err != nil {
			return errors.Wrap(err, "cannot add migration secret")
		}
	}

	_, err = r.AddServiceToManage(ctx, service)
	if err != nil {
		return errors.Wrapf(err, "cannot add service %s", service.GetName())
	}

	configMap, err := r.GetConfigMap(ctx, notaryserver)
	if err != nil {
		return errors.Wrap(err, "cannot get configMap")
	}

	configMapResource, err := r.AddConfigMapToManage(ctx, configMap, storageSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add configMap %s", configMap.GetName())
	}

	deployment, err := r.GetDeployment(ctx, notaryserver)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	_, err = r.AddDeploymentToManage(ctx, deployment, configMapResource, migrationSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add deployment %s", deployment.GetName())
	}

	return nil
}

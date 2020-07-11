package harbor

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.Harbor{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error { // nolint:funlen
	harbor, ok := resource.(*goharborv1alpha2.Harbor)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	registryAuthSecret, registryHTTPSecret, err := r.AddRegistryConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add registry configuration")
	}

	coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey, err := r.AddCoreConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	jobServiceSecret, err := r.AddJobServiceConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	chartMuseumAuthSecret, err := r.AddChartMuseumConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	_, err = r.AddNotaryServerConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	registry, err := r.AddRegistry(ctx, harbor, registryAuthSecret, registryHTTPSecret)
	if err != nil {
		return errors.Wrap(err, "cannot add registry")
	}

	_, err = r.AddRegistryController(ctx, harbor, registry)
	if err != nil {
		return errors.Wrap(err, "cannot add registry")
	}

	core, err := r.AddCore(ctx, harbor, registryAuthSecret, chartMuseumAuthSecret, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey)
	if err != nil {
		return errors.Wrap(err, "cannot add core")
	}

	_, err = r.AddJobService(ctx, harbor, core, coreSecret, jobServiceSecret)
	if err != nil {
		return errors.Wrap(err, "cannot add jobservice")
	}

	portal, err := r.AddPortal(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add portal")
	}

	_, err = r.AddChartMuseum(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add chartmuseum")
	}

	notaryServer, err := r.AddNotaryServer(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add notaryserver")
	}

	_, err = r.AddCoreIngress(ctx, harbor, core, portal, registry)
	if err != nil {
		return errors.Wrap(err, "cannot add core ingress")
	}

	_, err = r.AddNotaryIngress(ctx, harbor, notaryServer)
	if err != nil {
		return errors.Wrap(err, "cannot add notary ingress")
	}

	return nil
}

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

	_, _, internalTLSIssuer, err := r.AddInternalTLSIssuer(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add internal TLS issuer")
	}

	registryCertificate, registryAuthSecret, registryHTTPSecret, err := r.AddRegistryConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrap(err, "cannot add registry configuration")
	}

	coreCertificate, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey, err := r.AddCoreConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	jobServiceCertificate, jobServiceSecret, err := r.AddJobServiceConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	chartMuseumCertificate, chartMuseumAuthSecret, err := r.AddChartMuseumConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	_, err = r.AddNotaryServerConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add core configuration")
	}

	registry, err := r.AddRegistry(ctx, harbor, registryCertificate, registryAuthSecret, registryHTTPSecret)
	if err != nil {
		return errors.Wrap(err, "cannot add registry")
	}

	_, _, err = r.AddRegistryController(ctx, harbor, registry, internalTLSIssuer)
	if err != nil {
		return errors.Wrap(err, "cannot add registry")
	}

	core, err := r.AddCore(ctx, harbor, coreCertificate, registryAuthSecret, chartMuseumAuthSecret, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey)
	if err != nil {
		return errors.Wrap(err, "cannot add core")
	}

	_, err = r.AddJobService(ctx, harbor, core, jobServiceCertificate, coreSecret, jobServiceSecret)
	if err != nil {
		return errors.Wrap(err, "cannot add jobservice")
	}

	_, portal, err := r.AddPortal(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrap(err, "cannot add portal")
	}

	_, err = r.AddChartMuseum(ctx, harbor, chartMuseumCertificate)
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

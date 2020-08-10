package harbor

import (
	"context"

	"github.com/pkg/errors"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers"
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

	_, _, internalTLSIssuer, err := r.AddInternalTLSConfiguration(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot add internal TLS configuration")
	}

	registryCertificate, registryAuthSecret, registryHTTPSecret, err := r.AddRegistryConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s configuration", controllers.Registry)
	}

	coreCertificate, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey, err := r.AddCoreConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s configuration", controllers.Core)
	}

	jobServiceCertificate, jobServiceSecret, err := r.AddJobServiceConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s configuration", controllers.JobService)
	}

	chartMuseumCertificate, err := r.AddChartMuseumConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s configuration", controllers.ChartMuseum)
	}

	notaryCertIssuer, notarySignerCertificate, encryptionKey, notarySignerMigrationSecret, err := r.AddNotarySignerConfigurations(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s configuration", controllers.NotarySigner)
	}

	notaryAuthCert, notaryServerCertificate, notaryServerMigrationSecret, err := r.AddNotaryServerConfigurations(ctx, harbor, internalTLSIssuer, notaryCertIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s configuration", controllers.NotaryServer)
	}

	registry, err := r.AddRegistry(ctx, harbor, registryCertificate, registryAuthSecret, registryHTTPSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.Registry)
	}

	_, _, err = r.AddRegistryController(ctx, harbor, registry, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.RegistryController)
	}

	core, err := r.AddCore(ctx, harbor, coreCertificate, registryAuthSecret, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.Core)
	}

	_, err = r.AddJobService(ctx, harbor, core, jobServiceCertificate, coreSecret, jobServiceSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.JobService)
	}

	_, portal, err := r.AddPortal(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.Portal)
	}

	_, err = r.AddChartMuseum(ctx, harbor, chartMuseumCertificate, coreSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.ChartMuseum)
	}

	notaryServer, err := r.AddNotaryServer(ctx, harbor, notaryServerCertificate, notaryAuthCert, notaryServerMigrationSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.NotaryServer)
	}

	_, err = r.AddNotarySigner(ctx, harbor, notarySignerCertificate, encryptionKey, notarySignerMigrationSecret)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s", controllers.NotarySigner)
	}

	_, err = r.AddCoreIngress(ctx, harbor, core, portal, registry)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s ingress", controllers.Core)
	}

	_, err = r.AddNotaryIngress(ctx, harbor, notaryServer)
	if err != nil {
		return errors.Wrapf(err, "cannot add %s ingress", controllers.NotaryServer)
	}

	return nil
}

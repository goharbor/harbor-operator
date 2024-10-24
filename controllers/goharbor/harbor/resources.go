package harbor

import (
	"context"

	"github.com/pkg/errors"
	goharborv1 "github.com/plotly/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/plotly/harbor-operator/controllers"
	serrors "github.com/plotly/harbor-operator/pkg/controller/errors"
	"github.com/plotly/harbor-operator/pkg/resources"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1.Harbor{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error { //nolint:funlen
	harbor, ok := resource.(*goharborv1.Harbor)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	_, _, internalTLSIssuer, err := r.AddInternalTLSConfiguration(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "add internal TLS configuration")
	}

	registryCertificate, registryAuthSecret, registryHTTPSecret, err := r.AddRegistryConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s configuration", controllers.Registry)
	}

	coreCertificate, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey, err := r.AddCoreConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s configuration", controllers.Core)
	}

	jobServiceCertificate, jobServiceSecret, err := r.AddJobServiceConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s configuration", controllers.JobService)
	}

	trivyCertificate, trivyUpdateSecret, err := r.AddTrivyConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s configuration", controllers.Trivy)
	}

	registry, err := r.AddRegistry(ctx, harbor, registryCertificate, registryAuthSecret, registryHTTPSecret)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.Registry)
	}

	_, _, err = r.AddRegistryController(ctx, harbor, registry, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.RegistryController)
	}

	core, err := r.AddCore(ctx, harbor, coreCertificate, registryAuthSecret, coreCSRF, coreTokenCertificate, coreSecret, coreAdminPassword, coreEncryptionKey)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.Core)
	}

	_, err = r.AddJobService(ctx, harbor, core, jobServiceCertificate, coreSecret, jobServiceSecret)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.JobService)
	}

	_, portal, err := r.AddPortal(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.Portal)
	}

	exporterCertificate, err := r.AddExporterConfigurations(ctx, harbor, internalTLSIssuer)
	if err != nil {
		return errors.Wrapf(err, "add %s configuration", controllers.Exporter)
	}

	_, err = r.AddExporter(ctx, harbor, core, exporterCertificate)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.Exporter)
	}

	_, err = r.AddTrivy(ctx, harbor, trivyCertificate, trivyUpdateSecret)
	if err != nil {
		return errors.Wrapf(err, "add %s", controllers.Trivy)
	}

	_, err = r.AddCoreIngress(ctx, harbor, core, portal)
	if err != nil {
		return errors.Wrapf(err, "add %s ingress", controllers.Core)
	}

	err = r.AddNetworkPolicies(ctx, harbor)
	if err != nil {
		return errors.Wrapf(err, "add network policies")
	}

	return nil
}

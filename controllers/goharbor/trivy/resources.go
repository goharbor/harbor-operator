package trivy

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &goharborv1alpha2.Trivy{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	// Fetch the trivy resources definition
	trivy, ok := resource.(*goharborv1alpha2.Trivy)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	// Add Trivy service resources with a public port for Trivy deployment
	err := r.AddService(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot add service")
	}

	// Add Trivy config map resources with all Trivy custom definition
	err = r.AddConfigMap(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot add config map")
	}

	// Add Trivy secret resources with creds for Redis
	err = r.AddSecret(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot add secret")
	}

	// Add Trivy deployment and volumes resources
	err = r.AddDeployment(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot add deployment")
	}

	return nil
}

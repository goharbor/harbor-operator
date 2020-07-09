package v1alpha2

import (
	"context"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var notarysignerlog = logf.Log.WithName("notarysigner-resource")

const (
	NotarySignerMigrationSourceConfigKey = "notary-signer-migration-source"
)

var defaultNotarySignerMigrationSource string

func (r *NotarySigner) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	value, err := configstore.GetItemValue(NotarySignerMigrationSourceConfigKey)
	if err == nil {
		return errors.Wrap(err, "cannot get default migration source")
	}

	defaultNotarySignerMigrationSource = value

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha2-notarysigner,mutating=true,failurePolicy=ignore,groups=goharbor.io,resources=notarysigners,verbs=create;update,versions=v1alpha2,name=mnotarysigner.kb.io

var _ webhook.Defaulter = &NotarySigner{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *NotarySigner) Default() {
	notarysignerlog.Info("default", "name", r.Name)

	if r.Spec.Migration != nil {
		if r.Spec.Migration.DSN == "" {
			r.Spec.Migration.OpacifiedDSN = OpacifiedDSN{
				DSN: defaultNotarySignerMigrationSource,
			}
		}
	}
}

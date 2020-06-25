package v1alpha2 // nolint:dupl

import (
	"context"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

// +kubebuilder:webhook:path=/mutate-goharbor-io-goharbor-io-v1alpha2-notarysigner,mutating=true,failurePolicy=ignore,groups=goharbor.io.goharbor.io,resources=notarysigners,verbs=create;update,versions=v1alpha2,name=mnotarysigner.kb.io

var _ webhook.Defaulter = &NotarySigner{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *NotarySigner) Default() {
	notarysignerlog.Info("default", "name", r.Name)

	if !r.Spec.Migration.Disabled {
		if r.Spec.Migration.Source.DSN == "" {
			r.Spec.Migration.Source = OpacifiedDSN{
				DSN: defaultNotarySignerMigrationSource,
			}
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-goharbor-io-v1alpha2-notarysigner,mutating=false,failurePolicy=fail,groups=goharbor.io.goharbor.io,resources=notarysigners,versions=v1alpha2,name=vnotarysigner.kb.io

var _ webhook.Validator = &NotarySigner{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *NotarySigner) ValidateCreate() error {
	notarysignerlog.Info("validate create", "name", r.Name)

	return r.Spec.Migration.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *NotarySigner) ValidateUpdate(old runtime.Object) error {
	notarysignerlog.Info("validate update", "name", r.Name)

	return r.Spec.Migration.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *NotarySigner) ValidateDelete() error {
	notarysignerlog.Info("validate delete", "name", r.Name)

	return nil
}

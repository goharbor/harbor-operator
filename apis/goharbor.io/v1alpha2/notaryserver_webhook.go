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
var notaryserverlog = logf.Log.WithName("notaryserver-resource")

const (
	NotaryServerMigrationSourceConfigKey = "notary-server-migration-source"
)

var defaultNotaryServerMigrationSource string

func (r *NotaryServer) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	value, err := configstore.GetItemValue(NotaryServerMigrationSourceConfigKey)
	if err == nil {
		return errors.Wrap(err, "cannot get default migration source")
	}

	defaultNotaryServerMigrationSource = value

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-goharbor-io-goharbor-io-v1alpha2-notaryserver,mutating=true,failurePolicy=ignore,groups=goharbor.io.goharbor.io,resources=notaryservers,verbs=create;update,versions=v1alpha2,name=mnotaryserver.kb.io

var _ webhook.Defaulter = &NotaryServer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *NotaryServer) Default() {
	notaryserverlog.Info("default", "name", r.Name)

	if !r.Spec.Migration.Disabled {
		if r.Spec.Migration.Source.DSN == "" {
			r.Spec.Migration.Source = OpacifiedDSN{
				DSN: defaultNotaryServerMigrationSource,
			}
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-goharbor-io-v1alpha2-notaryserver,mutating=false,failurePolicy=fail,groups=goharbor.io.goharbor.io,resources=notaryservers,versions=v1alpha2,name=vnotaryserver.kb.io

var _ webhook.Validator = &NotaryServer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *NotaryServer) ValidateCreate() error {
	notaryserverlog.Info("validate create", "name", r.Name)

	return r.Spec.Migration.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *NotaryServer) ValidateUpdate(old runtime.Object) error {
	notaryserverlog.Info("validate update", "name", r.Name)

	return r.Spec.Migration.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *NotaryServer) ValidateDelete() error {
	notaryserverlog.Info("validate delete", "name", r.Name)

	return nil
}

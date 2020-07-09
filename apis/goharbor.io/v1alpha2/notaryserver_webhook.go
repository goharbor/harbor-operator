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

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha2-notaryserver,mutating=true,failurePolicy=ignore,groups=goharbor.io,resources=notaryservers,verbs=create;update,versions=v1alpha2,name=mnotaryserver.kb.io

var _ webhook.Defaulter = &NotaryServer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *NotaryServer) Default() {
	notaryserverlog.Info("default", "name", r.Name)

	if r.Spec.Migration != nil {
		if r.Spec.Migration.DSN == "" {
			r.Spec.Migration.OpacifiedDSN = OpacifiedDSN{
				DSN: defaultNotaryServerMigrationSource,
			}
		}
	}
}

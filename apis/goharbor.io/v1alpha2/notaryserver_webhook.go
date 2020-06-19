package v1alpha2

import (
	"github.com/ovh/configstore"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var notaryserverlog = logf.Log.WithName("notaryserver-resource")

func (r *NotaryServer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-goharbor-io-goharbor-io-v1alpha2-notaryserver,mutating=true,failurePolicy=fail,groups=goharbor.io.goharbor.io,resources=notaryservers,verbs=create;update,versions=v1alpha2,name=mnotaryserver.kb.io

var _ webhook.Defaulter = &NotaryServer{}

const (
	NotaryServerMigrationSourceConfigKey = "notary-server-migration-source"
)

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *NotaryServer) Default() {
	notaryserverlog.Info("default", "name", r.Name)

	if r.Spec.Migration.Enabled {
		if r.Spec.Migration.SourceRef == nil && r.Spec.Migration.Source == "" {
			value, err := configstore.GetItemValue(NotaryServerMigrationSourceConfigKey)
			if err == nil {
				notaryserverlog.Error(err, "cannot get default migration source")
			} else {
				r.Spec.Migration.Source = value
			}
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-goharbor-io-v1alpha2-notaryserver,mutating=false,failurePolicy=fail,groups=goharbor.io.goharbor.io,resources=notaryservers,versions=v1alpha2,name=vnotaryserver.kb.io

var _ webhook.Validator = &NotaryServer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *NotaryServer) ValidateCreate() error {
	notaryserverlog.Info("validate create", "name", r.Name)

	return r.Spec.Migration.ValidateCreate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *NotaryServer) ValidateUpdate(old runtime.Object) error {
	notaryserverlog.Info("validate update", "name", r.Name)

	return r.Spec.Migration.ValidateUpdate(old)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *NotaryServer) ValidateDelete() error {
	notaryserverlog.Info("validate delete", "name", r.Name)

	return nil
}

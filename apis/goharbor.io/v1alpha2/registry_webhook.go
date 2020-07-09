package v1alpha2

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var registrylog = logf.Log.WithName("registry-resource")

func (r *Registry) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-registry,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=registries,versions=v1alpha2,name=vregistry.kb.io

var _ webhook.Validator = &Registry{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Registry) ValidateCreate() error {
	registrylog.Info("validate create", "name", r.Name)

	return r.Spec.Storage.Driver.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Registry) ValidateUpdate(old runtime.Object) error {
	registrylog.Info("validate update", "name", r.Name)

	return r.Spec.Storage.Driver.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Registry) ValidateDelete() error {
	registrylog.Info("validate delete", "name", r.Name)

	return nil
}

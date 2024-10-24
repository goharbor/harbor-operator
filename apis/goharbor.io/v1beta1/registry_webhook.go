package v1beta1

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var registrylog = logf.Log.WithName("registry-resource")

func (r *Registry) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1beta1-registry,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=registries,versions=v1beta1,name=vregistry.kb.io,admissionReviewVersions={"v1beta1","v1"},sideEffects=None

var _ webhook.Validator = &Registry{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Registry) ValidateCreate() (admission.Warnings, error) {
	registrylog.Info("validate create", "name", r.Name)

	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Registry) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	registrylog.Info("validate update", "name", r.Name)

	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Registry) ValidateDelete() (admission.Warnings, error) {
	registrylog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func (r *Registry) Validate() (admission.Warnings, error) {
	var allErrs field.ErrorList

	err := r.Spec.Storage.Driver.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("storage").Child("driver"), r.Spec.Storage.Driver, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Registry"}, r.Name, allErrs)
}

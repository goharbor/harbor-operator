package v1alpha2

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var trivyLog = logf.Log.WithName("trivy-resource")

func (r *Trivy) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-trivy,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=trivies,versions=v1alpha2,name=mtrivy.kb.io

var _ webhook.Validator = &Trivy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Trivy) ValidateCreate() error {
	registrylog.Info("validate create", "name", r.Name)

	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Trivy) ValidateUpdate(old runtime.Object) error {
	registrylog.Info("validate update", "name", r.Name)

	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Trivy) ValidateDelete() error {
	registrylog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *Trivy) Validate() error {
	var allErrs field.ErrorList

	errs := r.Spec.Server.Validate()
	if len(errs) == 0 {
		return nil
	}

	for fieldName, err := range errs {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("server").Child(fieldName), r.Spec.Server, err.Error()))
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Trivy"}, r.Name, allErrs)
}

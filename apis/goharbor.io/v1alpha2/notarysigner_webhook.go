package v1alpha2 // nolint:dupl

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
var notarysignerlog = logf.Log.WithName("notarysigner-resource")

func (n *NotarySigner) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(n).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-notarysigner,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=notarysigners,versions=v1alpha2,name=vnotarysigner.kb.io

var _ webhook.Validator = &NotarySigner{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (n *NotarySigner) ValidateCreate() error {
	notarysignerlog.Info("validate create", "name", n.Name)

	return n.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (n *NotarySigner) ValidateUpdate(old runtime.Object) error {
	notarysignerlog.Info("validate update", "name", n.Name)

	return n.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (n *NotarySigner) ValidateDelete() error {
	notarysignerlog.Info("validate delete", "name", n.Name)

	return nil
}

func (n *NotarySigner) Validate() error {
	var allErrs field.ErrorList

	err := n.Spec.Migration.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("migration"), n.Spec.Migration, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "NotarySigner"}, n.Name, allErrs)
}

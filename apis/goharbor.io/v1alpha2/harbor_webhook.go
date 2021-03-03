package v1alpha2

import (
	"context"
	"fmt"
	"net/url"

	"github.com/goharbor/harbor-operator/pkg/version"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var harborlog = logf.Log.WithName("harbor-resource")

func (h *Harbor) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(h).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha2-harbor,mutating=true,failurePolicy=fail,groups=goharbor.io,resources=harbors,verbs=create;update,versions=v1alpha2,name=mharbor.kb.io,admissionReviewVersions={"v1beta1"},sideEffects=None

var _ webhook.Defaulter = &Harbor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (h *Harbor) Default() {
	if h.Spec.Version == "" {
		h.Spec.Version = version.Default()
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-harbor,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=harbors,versions=v1alpha2,name=vharbor.kb.io,admissionReviewVersions={"v1beta1"},sideEffects=None

var _ webhook.Validator = &Harbor{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (h *Harbor) ValidateCreate() error {
	harborlog.Info("validate create", "name", h.Name)

	return h.Validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (h *Harbor) ValidateUpdate(old runtime.Object) error {
	harborlog.Info("validate update", "name", h.Name)

	obj, ok := old.(*Harbor)
	if !ok {
		return field.Invalid(field.NewPath("spec"), old, fmt.Sprintf("failed type assertion to Harbor on kind: %s", old.GetObjectKind().GroupVersionKind().String()))
	}

	return h.Validate(obj)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (h *Harbor) ValidateDelete() error {
	harborlog.Info("validate delete", "name", h.Name)

	return nil
}

func (h *Harbor) Validate(old *Harbor) error {
	var allErrs field.ErrorList

	err := h.Spec.ImageChartStorage.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("imageChartStorage"), h.Spec.ImageChartStorage, err.Error()))
	}

	_, err = url.Parse(h.Spec.ExternalURL)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("externalURL"), h.Spec.ExternalURL, err.Error()))
	}

	if h.Spec.Database == nil {
		allErrs = append(allErrs, required(field.NewPath("spec").Child("database")))
	}

	if h.Spec.Redis == nil {
		allErrs = append(allErrs, required(field.NewPath("spec").Child("redis")))
	}

	if old == nil { // create harbor resource
		if err := version.Validate(h.Spec.Version); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("version"), h.Spec.Version, err.Error()))
		}
	} else { // update harbor resource
		if err := version.UpgradeAllowed(old.Spec.Version, h.Spec.Version); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("version"), h.Spec.Version, err.Error()))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Harbor"}, h.Name, allErrs)
}

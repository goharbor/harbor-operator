package v1beta1

import (
	"context"

	"github.com/pkg/errors"
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
var hplog = logf.Log.WithName("harborproject-resource")

func (hp *HarborProject) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(hp).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1beta1-harborproject,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=harborprojects,versions=v1beta1,name=vharborproject.kb.io,admissionReviewVersions={"v1beta1","v1"},sideEffects=None

var _ webhook.Validator = &HarborProject{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (hp *HarborProject) ValidateCreate() (admission.Warnings, error) {
	hplog.Info("validate create", "name", hp.Name)

	return hp.Validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (hp *HarborProject) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	hplog.Info("validate update", "name", hp.Name)

	obj, ok := old.(*HarborProject)
	if !ok {
		return nil, errors.Errorf("failed type assertion on kind: %s", old.GetObjectKind().GroupVersionKind().String())
	}

	return hp.Validate(obj)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (hp *HarborProject) ValidateDelete() (admission.Warnings, error) {
	hplog.Info("validate delete", "name", hp.Name)

	return nil, nil
}

func (hp *HarborProject) Validate(old *HarborProject) (admission.Warnings, error) {
	var allErrs field.ErrorList

	if old != nil { // update harborproject resource
		if hp.Spec.ProjectName != old.Spec.ProjectName {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("projectName"), hp.Spec.ProjectName, "field cannot be changed after initial creation"))
		}

		if hp.Spec.HarborServerConfig != old.Spec.HarborServerConfig {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("harborServerConfig"), hp.Spec.HarborServerConfig, "field cannot be changed after initial creation"))
		}
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "HarborProject"}, hp.Name, allErrs)
}

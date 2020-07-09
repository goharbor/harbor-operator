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
var jobservicelog = logf.Log.WithName("jobservice-resource")

func (js *JobService) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(js).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha2-jobservice,mutating=true,failurePolicy=fail,groups=goharbor.io,resources=jobservices,verbs=create;update,versions=v1alpha2,name=mjobservice.kb.io

var _ webhook.Defaulter = &JobService{}

// Default implements webhook.Defaulter, so a webhook will be registered for the type.
func (js *JobService) Default() {
	jobservicelog.Info("default", "name", js.Name)
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-jobservice,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=jobservices,versions=v1alpha2,name=vjobservice.kb.io

var _ webhook.Validator = &JobService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (js *JobService) ValidateCreate() error {
	jobservicelog.Info("validate create", "name", js.Name)

	return js.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (js *JobService) ValidateUpdate(old runtime.Object) error {
	jobservicelog.Info("validate update", "name", js.Name)

	return js.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (js *JobService) ValidateDelete() error {
	jobservicelog.Info("validate delete", "name", js.Name)

	return nil
}

func (js *JobService) Validate() error {
	var allErrs field.ErrorList

	err := js.Spec.JobLoggers.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("jobLoggers"), js.Spec.JobLoggers, err.Error()))
	}

	err = js.Spec.Loggers.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("loggers"), js.Spec.Loggers, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "JobService"}, js.Name, allErrs)
}

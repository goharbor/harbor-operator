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
)

// log is for logging in this package.
var jobservicelog = logf.Log.WithName("jobservice-resource")

func (jobservice *JobService) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(jobservice).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1beta1-jobservice,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=jobservices,versions=v1beta1,name=vjobservice.kb.io,admissionReviewVersions={"v1beta1","v1"},sideEffects=None

var _ webhook.Validator = &JobService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (jobservice *JobService) ValidateCreate() error {
	jobservicelog.Info("validate create", "name", jobservice.Name)

	return jobservice.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (jobservice *JobService) ValidateUpdate(old runtime.Object) error {
	jobservicelog.Info("validate update", "name", jobservice.Name)

	return jobservice.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (jobservice *JobService) ValidateDelete() error {
	jobservicelog.Info("validate delete", "name", jobservice.Name)

	return nil
}

func (jobservice *JobService) Validate() error {
	var allErrs field.ErrorList

	err := jobservice.Spec.JobLoggers.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("jobLoggers"), jobservice.Spec.JobLoggers, err.Error()))
	}

	err = jobservice.Spec.Loggers.Validate()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("loggers"), jobservice.Spec.Loggers, err.Error()))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "JobService"}, jobservice.Name, allErrs)
}

package v1alpha2

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var jobservicelog = logf.Log.WithName("jobservice-resource")

func (js *JobService) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(js).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-containerregistry-ovhcloud-com-v1alpha2-jobservice,mutating=true,failurePolicy=fail,groups=containerregistry.ovhcloud.com,resources=jobservices,verbs=create;update,versions=v1alpha2,name=mjobservice.kb.io

var _ webhook.Defaulter = &Harbor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (js *JobService) Default() {
	jobservicelog.Info("default", "name", js.Name)

	if js.Spec.WorkerCount == 0 {
		js.Spec.WorkerCount = 3
	}
}

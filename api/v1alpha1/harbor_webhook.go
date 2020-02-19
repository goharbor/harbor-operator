package v1alpha1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var harborlog = logf.Log.WithName("harbor-resource")

func (r *Harbor) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-containerregistry-ovhcloud-com-v1alpha1-harbor,mutating=true,failurePolicy=fail,groups=containerregistry.ovhcloud.com,resources=harbors,verbs=create;update,versions=v1alpha1,name=mharbor.kb.io

var _ webhook.Defaulter = &Harbor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Harbor) Default() {
	harborlog.Info("default", "name", r.Name)

	if r.Spec.Components.JobService != nil {
		if r.Spec.Components.JobService.WorkerCount == 0 {
			r.Spec.Components.JobService.WorkerCount = 3
		}
	}

	if r.Spec.HarborVersion == "" {
		r.Spec.HarborVersion = "1.10.0"
	}
}

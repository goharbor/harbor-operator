package v1alpha1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var harborlog = logf.Log.WithName("harbor-resource")

func (h *Harbor) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(h).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha1-harbor,mutating=true,failurePolicy=fail,groups=goharbor.io,resources=harbors,verbs=create;update,versions=v1alpha1,name=mharbor.kb.io

var _ webhook.Defaulter = &Harbor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (h *Harbor) Default() {
	harborlog.Info("default", "name", h.Name)

	if h.Spec.Components.JobService != nil {
		if h.Spec.Components.JobService.WorkerCount == 0 {
			h.Spec.Components.JobService.WorkerCount = 3
		}
	}

	if h.Spec.HarborVersion == "" {
		h.Spec.HarborVersion = "1.10.0"
	}
}

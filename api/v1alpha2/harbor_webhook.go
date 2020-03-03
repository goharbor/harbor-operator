package v1alpha2

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

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha2-harbor,mutating=true,failurePolicy=fail,groups=goharbor-io,resources=harbors,verbs=create;update,versions=v1alpha2,name=mharbor.kb.io

var _ webhook.Defaulter = &Harbor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Harbor) Default() {
	harborlog.Info("default", "name", r.Name)

	if r.Spec.HarborVersion == "" {
		r.Spec.HarborVersion = "1.10.0"
	}
}

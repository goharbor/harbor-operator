package v1beta1

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *RegistryController) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

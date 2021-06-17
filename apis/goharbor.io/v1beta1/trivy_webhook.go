package v1beta1

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func (t *Trivy) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(t).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

package v1beta1

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func (j *JobService) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(j).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

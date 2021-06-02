package v1beta1

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var notaryserverlog = logf.Log.WithName("notaryserver-resource")

func (r *NotaryServer) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

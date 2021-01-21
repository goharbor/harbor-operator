package v1alpha2 // nolint:dupl

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var notaryserverlog = logf.Log.WithName("notaryserver-resource")

func (n *NotaryServer) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(n).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-notaryserver,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=notaryservers,versions=v1alpha2,name=vnotaryserver.kb.io,admissionReviewVersions={"v1beta1"},sideEffects=None

var _ webhook.Validator = &NotaryServer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (n *NotaryServer) ValidateCreate() error {
	notaryserverlog.Info("validate create", "name", n.Name)

	return n.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (n *NotaryServer) ValidateUpdate(old runtime.Object) error {
	notaryserverlog.Info("validate update", "name", n.Name)

	return n.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (n *NotaryServer) ValidateDelete() error {
	notaryserverlog.Info("validate delete", "name", n.Name)

	return nil
}

func (n *NotaryServer) Validate() error {
	return nil
}

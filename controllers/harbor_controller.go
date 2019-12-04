/*
.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
)

// HarborReconciler reconciles a Harbor object
type HarborReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=harbors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=harbors/status,verbs=get;update;patch

func (r *HarborReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("harbor", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *HarborReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&containerregistryv1alpha1.Harbor{}).
		Complete(r)
}

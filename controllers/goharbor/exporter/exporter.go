package exporter

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// Reconciler reconciles a Exporter object.
type Reconciler struct {
	*commonCtrl.Controller
}

// +kubebuilder:rbac:groups=goharbor.io,resources=exporters,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=exporters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "setup common controller")
	}

	className, err := r.GetClassName(ctx)
	if err != nil {
		return errors.Wrap(err, "classname")
	}

	concurrentReconcile, err := config.GetInt(r.ConfigStore, config.ReconciliationKey, config.DefaultConcurrentReconcile)
	if err != nil {
		return errors.Wrap(err, "get concurrent reconcile")
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: className,
		}).
		For(r.NewEmpty(ctx)).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&netv1.NetworkPolicy{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrentReconcile,
		}).
		Complete(r)
}

func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, controllers.Exporter, r, configStore)

	return r, nil
}

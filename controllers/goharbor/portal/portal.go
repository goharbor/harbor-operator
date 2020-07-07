package portal

import (
	"context"
	"time"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
)

const (
	DefaultRequeueWait = 2 * time.Second
	ConfigImageKey     = "docker-image"
	DefaultImage       = "goharbor/harbor-portal:v2.0.0"
)

// Reconciler reconciles a Portal object.
type Reconciler struct {
	*commonCtrl.Controller
}

// +kubebuilder:rbac:groups=goharbor.io,resources=portals,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=portals/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	className, err := r.ConfigStore.GetItemValue(config.HarborClassKey)
	if err != nil {
		return errors.Wrap(err, "cannot get harbor class")
	}

	concurrentReconcile, err := r.ConfigStore.GetItemValueInt(config.ReconciliationKey)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: className,
		}).
		For(r.NewEmpty(ctx)).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: int(concurrentReconcile),
		}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

package portal

import (
	"context"
	"net/http"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	ConfigTemplatePathKey     = "template-path"
	DefaultConfigTemplatePath = "/etc/harbor-operator/portal-config.conf.tmpl"
	ConfigTemplateKey         = "template-content"
	ConfigImageKey            = "docker-image"
	DefaultImage              = config.DefaultRegistry + "goharbor/harbor-portal:v2.0.0"
)

// Reconciler reconciles a Portal object.
type Reconciler struct {
	*commonCtrl.Controller

	configError error
}

// +kubebuilder:rbac:groups=goharbor.io,resources=portals,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=portals/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	className, err := r.GetClassName(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get class name")
	}

	concurrentReconcile, err := r.ConfigStore.GetItemValueInt(config.ReconciliationKey)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	err = mgr.AddReadyzCheck(r.NormalizeName(ctx, "template"), func(req *http.Request) error { return r.configError })
	if err != nil {
		return errors.Wrap(err, "cannot add template ready check")
	}

	err = mgr.AddHealthzCheck(r.NormalizeName(ctx, "template"), func(req *http.Request) error { return r.configError })
	if err != nil {
		return errors.Wrap(err, "cannot add template health check")
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

func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	configTemplatePath, err := config.GetString(configStore, ConfigTemplatePathKey, DefaultConfigTemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "template path")
	}

	r := &Reconciler{
		configError: config.ErrNotReady,
	}

	configStore.FileCustomRefresh(configTemplatePath, func(data []byte) ([]configstore.Item, error) {
		r.configError = nil

		logger.Get(ctx).WithName("controller").WithName(controllers.Portal.String()).
			Info("config reloaded", "path", configTemplatePath)
		// TODO reconcile all core

		return []configstore.Item{configstore.NewItem(ConfigTemplateKey, string(data), config.DefaultPriority)}, nil
	})

	r.Controller = commonCtrl.NewController(ctx, controllers.Portal, r, configStore)

	return r, nil
}

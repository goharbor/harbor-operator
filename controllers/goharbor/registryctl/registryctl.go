package registryctl

import (
	"context"
	"net/http"
	"time"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	DefaultRequeueWait = 2 * time.Second

	ConfigTemplatePathKey     = "template-path"
	DefaultConfigTemplatePath = "/etc/harbor-operator/registryctl-config.yaml.tmpl"
	ConfigTemplateKey         = "template-content"
	ConfigImageKey            = "docker-image"
	DefaultImage              = "goharbor/harbor-registryctl:v2.0.0"
)

// Reconciler reconciles a RegistryController object.
type Reconciler struct {
	*commonCtrl.Controller
	registry.Reconciler

	configError error
}

// +kubebuilder:rbac:groups=goharbor.io,resources=registrycontrollers,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=registrycontrollers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;services,verbs=get;list;watch;create;update;patch;delete

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
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: int(concurrentReconcile),
		}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	configTemplatePath, err := configStore.GetItemValue(ConfigTemplatePathKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, errors.Wrap(err, "cannot get config template path")
		}

		configTemplatePath = DefaultConfigTemplatePath
	}

	r := &Reconciler{
		configError: config.ErrNotReady,
	}

	configStore.FileCustomRefresh(configTemplatePath, func(data []byte) ([]configstore.Item, error) {
		r.configError = nil

		logger.Get(ctx).WithName("controller").WithName(name).
			Info("config reloaded", "path", configTemplatePath)
		// TODO reconcile all core

		return []configstore.Item{configstore.NewItem(ConfigTemplateKey, string(data), config.DefaultPriority)}, nil
	})

	r.Reconciler.Controller = commonCtrl.NewController(ctx, controllers.Registry.String(), r, configStore)
	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

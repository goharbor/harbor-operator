package registry

import (
	"context"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	DefaultRequeueWait = 2 * time.Second
)

const (
	ConfigTemplatePathKey     = "template-path"
	DefaultConfigTemplatePath = "/etc/harbor-operator/registry-config.yaml.tmpl"
	ConfigTemplateKey         = "template-content"
	ConfigImageKey            = "docker-image"
	DefaultImage              = "goharbor/registry-photon:v2.0.0"
)

// Reconciler reconciles a Registry object.
type Reconciler struct {
	*commonCtrl.Controller
}

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=registries,verbs=get;list;watch
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=registries/status,verbs=get;update;patch

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
		For(&goharborv1alpha2.Registry{}).
		Owns(&appsv1.Deployment{}).
		Owns(&certv1.Certificate{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: int(concurrentReconcile),
		}).
		Complete(r)
}

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=registries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=registries/status,verbs=get;update;patch

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	configTemplatePath, err := configStore.GetItemValue(ConfigTemplatePathKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, errors.Wrap(err, "cannot get config template path")
		}

		configTemplatePath = DefaultConfigTemplatePath
	}

	l := logger.Get(ctx).WithName("controller").WithName(name)

	configStore.FileCustomRefresh(configTemplatePath, func(data []byte) ([]configstore.Item, error) {
		l.Info("config reloaded", "path", configTemplatePath)
		// TODO reconcile all registries
		return []configstore.Item{configstore.NewItem(ConfigTemplateKey, string(data), config.DefaultPriority)}, nil
	})

	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

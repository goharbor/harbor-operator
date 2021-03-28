package chartmuseum

import (
	"context"
	"time"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/config/template"
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

const (
	DefaultRequeueWait            = 2 * time.Second
	DefaultConfigTemplateFileName = "chartmuseum-config.yaml.tmpl"
)

// Reconciler reconciles a Chartmuseum object.
type Reconciler struct {
	*commonCtrl.Controller
}

// +kubebuilder:rbac:groups=goharbor.io,resources=chartmuseums,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=chartmuseums/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps;services,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	templateConfig, err := r.Template(ctx)
	if err != nil {
		return errors.Wrap(err, "template")
	}

	if err := mgr.AddReadyzCheck(r.NormalizeName(ctx, "template"), templateConfig.ReadyzCheck); err != nil {
		return errors.Wrap(err, "cannot add template ready check")
	}

	if err := mgr.AddHealthzCheck(r.NormalizeName(ctx, "template"), templateConfig.HealthzCheck); err != nil {
		return errors.Wrap(err, "cannot add template health check")
	}

	className, err := r.GetClassName(ctx)
	if err != nil {
		return errors.Wrap(err, "classname")
	}

	concurrentReconcile, err := config.GetInt(r.ConfigStore, config.ReconciliationKey, config.DefaultConcurrentReconcile)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: className,
		}).
		For(r.NewEmpty(ctx)).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&netv1.NetworkPolicy{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrentReconcile,
		}).
		Complete(r)
}

func (r *Reconciler) Template(ctx context.Context) (*template.ConfigTemplate, error) {
	templateConfig, err := template.FromConfigStore(r.ConfigStore, DefaultConfigTemplateFileName)
	if err != nil {
		return nil, errors.Wrap(err, "from configstore")
	}

	templateConfig.Register(r.ConfigStore)

	return templateConfig, nil
}

func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, controllers.ChartMuseum, r, configStore)

	return r, nil
}

package project

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/builder"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	finalizerID                  string = "harborproject.goharbor.io/finalizer"
	defaultRequeueAfterMinutes   int    = 5
	requeueAfterMinutesConfigKey string = "requeue-after-minutes"
)

// New HarborProject reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborProject, nil, configStore)

	return r, nil
}

// Reconciler reconciles a project cr.
type Reconciler struct {
	*commonCtrl.Controller
	Scheme              *runtime.Scheme
	Harbor              *v2.Client
	RequeueAfterMinutes int
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harborprojects/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harborprojects/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	concurrentReconcile, err := config.GetInt(r.ConfigStore, config.ReconciliationKey, config.DefaultConcurrentReconcile)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	requeueAfterMinutes, err := config.GetInt(r.ConfigStore, requeueAfterMinutesConfigKey, defaultRequeueAfterMinutes)
	if err != nil {
		return errors.Wrap(err, "cannot get requeue after config value")
	}

	r.RequeueAfterMinutes = requeueAfterMinutes
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return builder.ControllerManagedBy(mgr).
		For(&goharborv1.HarborProject{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrentReconcile,
		}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *Reconciler) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	suffixes = append([]string{"HarborProject"}, suffixes...)

	return strings.NormalizeName(name, suffixes...)
}

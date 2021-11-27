package trait

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	goharboriov1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/trait/webhook"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	harborstring "github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles a HarborClusterTrait object.
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	*commonCtrl.Controller
}

// New HarborCluster reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborClusterTrait, nil, configStore)

	return r, nil
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborclustertraits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harborclustertraits/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("harborclustertrait", req.NamespacedName)
	log.Info("start to reconcile.")

	var trait goharboriov1beta1.HarborClusterTrait
	if err := r.Get(ctx, req.NamespacedName, &trait); err != nil {
		log.Error(err, "unable to fetch HarborClusterTrait")

		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	for _, affinity := range trait.Spec.Affinities {
		for k, v := range affinity.Selector.MatchLabels {
			key := strings.Join([]string{k, v}, "=")

			if trait.DeletionTimestamp != nil {
				webhook.TraitMap.Delete(key)
				log.Info("success to remove key from trait_map.", "key", key)

				continue
			}

			webhook.TraitMap.Store(key, affinity.Affinity)
			log.Info("success to insert key from trait_map.", "key", key)
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Log = ctrl.Log.WithName("controllers").WithName("HarborClusterTrait")

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharboriov1beta1.HarborClusterTrait{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *Reconciler) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	suffixes = append([]string{"HarborClusterTrait"}, suffixes...)

	return harborstring.NormalizeName(name, suffixes...)
}

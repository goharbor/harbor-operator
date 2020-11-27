package harborcluster

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/cache"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/ovh/configstore"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconciler reconciles a HarborCluster object
type Reconciler struct {
	*commonCtrl.Controller

	CacheCtrl    lcm.Controller
	DatabaseCtrl lcm.Controller
	StorageCtrl  lcm.Controller
	HarborCtrl   *harbor.HarborController
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harborclusters/status,verbs=get;update;patch

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	dClient, err := k8s.NewDynamicClient()
	if err != nil {
		r.Log.Error(err, "unable to create dynamic client")
		return err
	}

	r.CacheCtrl = cache.NewRedisController(ctx,
		k8s.WithLog(r.Log),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(k8s.WrapDClient(dClient)),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())))
	r.DatabaseCtrl = database.NewDatabaseController(ctx,
		k8s.WithLog(r.Log),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(k8s.WrapDClient(dClient)),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())))
	r.StorageCtrl = storage.NewMinIOController(ctx,
		k8s.WithLog(r.Log),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(k8s.WrapDClient(dClient)),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())))
	r.HarborCtrl = harbor.NewHarborController(ctx,
		k8s.WithLog(r.Log),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(k8s.WrapDClient(dClient)),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())))

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1alpha2.HarborCluster{}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {

	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

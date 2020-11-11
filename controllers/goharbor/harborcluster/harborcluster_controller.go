package harborcluster

import (
	"context"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/cache"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/database"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/harbor"
	"github.com/goharbor/harbor-operator/pkg/cluster/storage"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/lcm"
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
	HarborCtrl   lcm.Controller
}

// +kubebuilder:rbac:groups=goharbor.goharbor.io,resources=harborclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.goharbor.io,resources=harborclusters/status,verbs=get;update;patch

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1alpha2.HarborCluster{}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {

	//dClient, err := k8s.NewDynamicClient()
	//if err != nil {
	//	return nil, err
	//}

	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	// TODO add args to service controller
	r.CacheCtrl = cache.NewCacheController()
	r.DatabaseCtrl = database.NewDatabaseController()
	r.StorageCtrl = storage.NewMinIOController()
	r.HarborCtrl = harbor.NewHarborController()

	return r, nil
}

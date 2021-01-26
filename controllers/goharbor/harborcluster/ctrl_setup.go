package harborcluster

import (
	"context"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"

	"github.com/go-logr/logr"
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/cache"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/ovh/configstore"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TODO: Refactor to inherit the common reconciler in future
// Reconciler reconciles a HarborCluster object.
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	// In case
	ConfigStore *configstore.Store
	Name        string
	Version     string

	CacheCtrl    lcm.Controller
	DatabaseCtrl lcm.Controller
	StorageCtrl  lcm.Controller
	HarborCtrl   *harbor.Controller
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=harborclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=acid.zalan.do,resources=postgresqls;operatorconfigurations,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=acid.zalan.do,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.spotahome.com,resources=redisfailovers,verbs=*
// +kubebuilder:rbac:groups=minio.min.io,resources=*,verbs=*
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets;deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harbors,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	dClient, err := k8s.NewDynamicClient()
	if err != nil {
		r.Log.Error(err, "unable to create dynamic client")

		return err
	}

	r.CacheCtrl = cache.NewRedisController(ctx,
		k8s.WithLog(r.Log.WithName("cache")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(k8s.WrapDClient(dClient)),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())),
		k8s.WithConfigStore(r.ConfigStore),
	)
	r.DatabaseCtrl = database.NewDatabaseController(ctx,
		k8s.WithLog(r.Log.WithName("database")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(k8s.WrapDClient(dClient)),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())),
		k8s.WithConfigStore(r.ConfigStore),
	)
	r.StorageCtrl = storage.NewMinIOController(ctx,
		k8s.WithLog(r.Log.WithName("storage")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())),
		k8s.WithConfigStore(r.ConfigStore),
	)
	r.HarborCtrl = harbor.NewHarborController(ctx,
		k8s.WithLog(r.Log.WithName("harbor")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithClient(k8s.WrapClient(ctx, mgr.GetClient())))

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1alpha2.HarborCluster{}).
		Owns(&goharborv1alpha2.Harbor{}).
		Owns(&redisOp.RedisFailover{}).
		Owns(&postgresv1.Postgresql{}).
		Owns(&minio.Tenant{}).
		WithEventFilter(harborClusterPredicateFuncs).
		Complete(r)
}

// New HarborCluster reconciler.
func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	return &Reconciler{
		Name:        name,
		Version:     application.GetVersion(ctx),
		ConfigStore: configStore,
		Log:         ctrl.Log.WithName(application.GetName(ctx)).WithName("controller").WithValues("controller", name),
	}, nil
}

// harborClusterPredicateFuncs do filter for events.
var harborClusterPredicateFuncs = predicate.Funcs{
	// we do not care other events
	UpdateFunc: func(event event.UpdateEvent) bool {
		old, ok := event.ObjectOld.(*goharborv1alpha2.HarborCluster)
		if !ok {
			return true
		}

		new, ok := event.ObjectNew.(*goharborv1alpha2.HarborCluster)
		if !ok {
			return true
		}
		// when status was not changed and spec was not changed, not need reconcile
		if equality.Semantic.DeepDerivative(old.Spec, new.Spec) && old.Status.Status == new.Status.Status {
			return false
		}

		return true
	},
}

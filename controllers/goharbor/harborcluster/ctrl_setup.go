package harborcluster

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/builder"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/cache"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/apis/minio.min.io/v2"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	minioCRD    = "tenants.minio.min.io"
	redisCRD    = "redisfailovers.databases.spotahome.com"
	postgresCRD = "postgresqls.acid.zalan.do"
)

// TODO: Refactor to inherit the common reconciler in future
// Reconciler reconciles a HarborCluster object.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// In case
	Name string

	CacheCtrl    lcm.Controller
	DatabaseCtrl lcm.Controller
	StorageCtrl  lcm.Controller
	HarborCtrl   *harbor.Controller

	*commonCtrl.Controller // TODO: move the Reconcile to pkg/controller.Controller
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
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	if err := r.SetupWithManager(ctx, mgr); err != nil {
		return err
	}

	concurrentReconcile, err := config.GetInt(r.ConfigStore, config.ReconciliationKey, config.DefaultConcurrentReconcile)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	dClient, err := k8s.DynamicClient()
	if err != nil {
		r.Log.Error(err, "unable to create dynamic client")

		return err
	}

	r.CacheCtrl = cache.NewRedisController(
		k8s.WithLog(r.Log.WithName("cache")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(dClient),
		k8s.WithClient(mgr.GetClient()),
		k8s.WithConfigStore(r.ConfigStore),
	)
	r.DatabaseCtrl = database.NewDatabaseController(
		k8s.WithLog(r.Log.WithName("database")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithDClient(dClient),
		k8s.WithClient(mgr.GetClient()),
		k8s.WithConfigStore(r.ConfigStore),
	)
	r.StorageCtrl = storage.NewMinIOController(
		k8s.WithLog(r.Log.WithName("storage")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithClient(mgr.GetClient()),
		k8s.WithConfigStore(r.ConfigStore),
	)
	r.HarborCtrl = harbor.NewHarborController(
		k8s.WithLog(r.Log.WithName("harbor")),
		k8s.WithScheme(mgr.GetScheme()),
		k8s.WithClient(mgr.GetClient()))

	return builder.ControllerManagedBy(mgr).
		For(&goharborv1.HarborCluster{}).
		Owns(&batchv1.Job{}).
		Owns(&goharborv1.Harbor{}).
		TryOwns(&minio.Tenant{}, minioCRD).
		TryOwns(&postgresv1.Postgresql{}, postgresCRD).
		TryOwns(&redisOp.RedisFailover{}, redisCRD).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrentReconcile,
		}).
		Complete(r)
}

func (r *Reconciler) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	suffixes = append([]string{"HarborCluster"}, suffixes...)

	return strings.NormalizeName(name, suffixes...)
}

// New HarborCluster reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborCluster, nil, configStore)

	return r, nil
}

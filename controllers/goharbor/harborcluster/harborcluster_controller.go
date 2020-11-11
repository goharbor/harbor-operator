package harborcluster

import (
	"context"

	"github.com/go-logr/logr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/goharbor/harbor-operator/pkg/k8s"

	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/cache"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/database"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/harbor"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/storage"
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

	scheme := newScheme()
	dClient, err := k8s.NewDynamicClient()
	if err != nil {
		return nil, err
	}

	client, err := k8s.NewClient(scheme)
	if err != nil {
		return nil, err
	}

	option := &k8s.GetOptions{
		CXT:     ctx,
		Client:  k8s.WrapClient(ctx, client),
		Log:     newLog(),
		DClient: k8s.WrapDClient(dClient),
		Scheme:  scheme,
	}

	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	// TODO add args to service controller
	r.CacheCtrl = cache.NewCacheController()
	r.DatabaseCtrl = database.NewDatabaseController(option)
	r.StorageCtrl = storage.NewMinIOController()
	r.HarborCtrl = harbor.NewHarborController()

	return r, nil
}

func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	return scheme
}

func newLog() logr.Logger {
	return ctrl.Log.WithName("controllers").WithName("HarborCluster")
}

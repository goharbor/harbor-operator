package harborcluster

import (
	"context"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/ovh/configstore"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconciler reconciles a HarborCluster object
type Reconciler struct {
	*commonCtrl.Controller

	ServiceGetter
}

// +kubebuilder:rbac:groups=goharbor.goharbor.io,resources=harborclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.goharbor.io,resources=harborclusters/status,verbs=get;update;patch

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1alpha2.HarborCluster{}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {

	dClient, err := k8s.NewDynamicClient()
	if err != nil {
		return nil, err
	}

	r := &Reconciler{}

	option := &GetOptions{
		Client:  k8s.WrapClient(ctx, r.Client),
		Log:     r.Log,
		DClient: k8s.WrapDClient(dClient),
		Scheme:  r.Scheme,
	}
	r.ServiceGetter = NewServiceGetterImpl(option)

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

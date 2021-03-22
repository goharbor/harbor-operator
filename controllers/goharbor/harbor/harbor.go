package harbor

import (
	"context"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	DefaultRequeueWait = 2 * time.Second
)

// Reconciler reconciles a Harbor object.
type Reconciler struct {
	*commonCtrl.Controller
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harbors,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=harbors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=goharbor.io,resources=chartmuseums;cores;jobservices;notaryservers;notarysigners;portals;registries;registrycontrollers;trivies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers;certificates,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	className, err := r.GetClassName(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get class name")
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
		Owns(&goharborv1.ChartMuseum{}).
		Owns(&goharborv1.Core{}).
		Owns(&goharborv1.JobService{}).
		Owns(&goharborv1.Portal{}).
		Owns(&goharborv1.Registry{}).
		Owns(&goharborv1.RegistryController{}).
		Owns(&goharborv1.NotaryServer{}).
		Owns(&goharborv1.NotarySigner{}).
		Owns(&corev1.Secret{}).
		Owns(&certv1.Issuer{}).
		Owns(&certv1.Certificate{}).
		Owns(&netv1beta1.Ingress{}).
		Owns(&netv1.NetworkPolicy{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrentReconcile,
		}).
		Complete(r)
}

func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, controllers.Harbor, r, configStore)

	return r, nil
}

package jobservice

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/controllers/config"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	DefaultRequeueWait = 2 * time.Second
)

// Reconciler reconciles a Harbor object
type Reconciler struct {
	*common.Controller

	Log logr.Logger

	Config config.Config
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := r.InitResources()
	if err != nil {
		return errors.Wrap(err, "cannot initialize resources")
	}

	err = r.Controller.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: r.Config.ClassName,
		}).
		For(&goharborv1alpha2.JobService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&certv1.Certificate{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&netv1.Ingress{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Config.ConcurrentReconciles,
		}).
		Complete(r)
}

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=jobservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=jobservices/status,verbs=get;update;patch

func New(ctx context.Context, name, version string, config *config.Config) (*Reconciler, error) {
	return &Reconciler{
		Controller: common.NewController(name, version),
		Log:        logger.Get(ctx).WithName("controller").WithName(name),
		Config:     *config,
	}, nil
}

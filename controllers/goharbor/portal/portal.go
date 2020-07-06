package portal

import (
	"context"
	"time"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
)

const (
	DefaultRequeueWait = 2 * time.Second
	ConfigImageKey     = "docker-image"
	DefaultImage       = "goharbor/harbor-portal:v2.0.0"
)

// Reconciler reconciles a Portal object.
type Reconciler struct {
	*commonCtrl.Controller
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	className, err := r.ConfigStore.GetItemValue(config.HarborClassKey)
	if err != nil {
		return errors.Wrap(err, "cannot get harbor class")
	}

	concurrentReconcile, err := r.ConfigStore.GetItemValueInt(config.ReconciliationKey)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: className,
		}).
		For(&goharborv1alpha2.Portal{}).
		Owns(&appsv1.Deployment{}).
		Owns(&certv1.Certificate{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&netv1.Ingress{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: int(concurrentReconcile),
		}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

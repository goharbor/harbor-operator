package harbor

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/logger"
)

const (
	DefaultRequeueWait = 2 * time.Second
)

type Config struct {
	ClassName            string
	ConcurrentReconciles int
	WatchChildren        bool
}

// Reconciler reconciles a Harbor object
type Reconciler struct {
	client.Client

	Name    string
	Version string

	Log    logr.Logger
	Scheme *runtime.Scheme

	RestConfig *rest.Config

	Config Config
}

func (r *Reconciler) GetVersion() string {
	return r.Version
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.RestConfig = mgr.GetConfig()

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(r.GetEventFilter()).
		For(&containerregistryv1alpha1.Harbor{}).
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

func New(ctx context.Context, name, version string, config *Config) (*Reconciler, error) {
	return &Reconciler{
		Name:    name,
		Version: version,
		Log:     logger.Get(ctx).WithName("controller").WithName(name),
		Config:  *config,
	}, nil
}

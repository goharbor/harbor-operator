package harborserverconfiguration

import (
	"context"
	"fmt"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	harborClient "github.com/goharbor/harbor-operator/pkg/rest"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultCycle = 5 * time.Minute
)

// New HarborServerConfiguration reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborServerConfiguration, nil, configStore)

	return r, nil
}

// Reconciler reconciles a HarborServerConfiguration object.
type Reconciler struct {
	*commonCtrl.Controller
	client.Client
	Scheme *runtime.Scheme
	Harbor *v2.Client
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborserverconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harborserverconfigurations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile the HarborServerConfiguration.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("harborserverconfiguration", req.NamespacedName)
	log.Info("Starting HarborServerConfiguration Reconciler")

	// Get the configuration first
	hsc := &goharborv1.HarborServerConfiguration{}
	if err := r.Client.Get(ctx, req.NamespacedName, hsc); err != nil {
		if apierr.IsNotFound(err) {
			// It could have been deleted after reconcile request coming in.
			log.Info("Harbor server configuration does not exist")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get HarborServerConfiguraiton error: %w", err)
	}

	defer func() {
		if err != nil {
			hsc.Status.Status = goharborv1.HarborServerConfigurationStatusFail
			hsc.Status.Message = err.Error()
		} else {
			hsc.Status.Status = goharborv1.HarborServerConfigurationStatusReady
			hsc.Status.Reason = ""
			hsc.Status.Message = ""
		}

		log.Info("Reconcile end", "result", res, "error", err, "updateStatusError", r.Client.Status().Update(ctx, hsc))
	}()

	// Create harbor client
	harborv2, err := harborClient.CreateHarborV2Client(ctx, r.Client, hsc)
	if err != nil {
		log.Error(err, "failed to create harbor client")

		return ctrl.Result{}, err
	}

	r.Harbor = harborv2.WithContext(ctx)

	// Check if the configuration is being deleted
	if !hsc.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Harbor server configuration is being deleted")

		return ctrl.Result{}, nil
	}

	// Check server health and construct status
	err = r.checkServerHealth()
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Finished HarborServerConfiguration Reconciler")
	// The health should be rechecked after a reasonable cycle
	return ctrl.Result{
		RequeueAfter: defaultCycle,
	}, nil
}

// SetupWithManager for HarborServerConfiguration reconcile controller.
func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1.HarborServerConfiguration{}).
		Complete(r)
}

func (r *Reconciler) checkServerHealth() error {
	errStr := ""

	healthPayload, err := r.Harbor.CheckHealth()
	if err != nil {
		r.Log.Error(err, "check harbor server health failed.")

		return err
	}

	for _, comp := range healthPayload.Components {
		if len(comp.Error) > 0 {
			errStr += "Component " + comp.Name + ": " + comp.Error
		}
	}

	if len(errStr) == 0 {
		return nil
	}

	return errors.Errorf(errStr)
}

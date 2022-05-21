package robotaccount

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	goharboriov1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	harborClient "github.com/goharbor/harbor-operator/pkg/rest"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	"github.com/ovh/configstore"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultCycle = 5 * time.Second
)

// RobotAccountReconciler reconciles a RobotAccount object.
type Reconciler struct {
	client.Client
	*commonCtrl.Controller
	Log    logr.Logger
	Scheme *runtime.Scheme
	Harbor *v2.Client
}

// New RobotAccount reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborCluster, nil, configStore)

	return r, nil
}

//+kubebuilder:rbac:groups=goharbor.io,resources=robotaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=goharbor.io,resources=robotaccounts/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RobotAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("RobotAccount", req.NamespacedName)
	log.Info("Starting RobotAccount Reconciler")

	// Get the robotaccount first
	ra := &goharboriov1beta1.RobotAccount{}
	if err := r.Client.Get(ctx, req.NamespacedName, ra); err != nil {
		if apierr.IsNotFound(err) {
			// It could have been deleted after reconcile request coming in.
			log.Info(fmt.Sprintf("Harbor robotaccount %s does not exist", req.NamespacedName))

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get HarborServerConfiguraiton error: %w", err)
	}

	// Check if the robotaccount is being deleted
	if !ra.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("robotaccount is being deleted", "name", ra.Name)

		return ctrl.Result{}, nil
	}

	harborCfg, err := r.getHarborServerConfig(ctx, ra.Spec.HarborServerConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error finding harborCfg: %w", err)
	}

	if harborCfg == nil {
		log.Info("no default hsc for ra: ", req.NamespacedName, ", skip RobotAccount creation")

		return ctrl.Result{RequeueAfter: defaultCycle}, nil
	}

	// Create harbor client
	harborv2, err := harborClient.CreateHarborV2Client(ctx, r.Client, harborCfg)
	if err != nil {
		log.Error(err, "failed to create harbor client")

		return ctrl.Result{}, err
	}

	r.Harbor = harborv2.WithContext(ctx)

	_, err = r.Harbor.CreateRobotAccount(ra)

	if err != nil {
		log.Error(err, "failed to create robot account")

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) getHarborServerConfig(ctx context.Context, name string) (*goharboriov1beta1.HarborServerConfiguration, error) {
	hsc := &goharboriov1beta1.HarborServerConfiguration{}
	// HarborServerConfiguration is cluster scoped resource
	namespacedName := types.NamespacedName{
		Name: name,
	}
	if err := r.Client.Get(ctx, namespacedName, hsc); err != nil {
		// Explicitly check not found error
		if apierr.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return hsc, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&goharboriov1beta1.RobotAccount{}).
		Complete(r)
}

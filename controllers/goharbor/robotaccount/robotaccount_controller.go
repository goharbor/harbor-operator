package robotaccount

import (
	"context"
	"errors"
	"fmt"

	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/robot"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	harborClient "github.com/goharbor/harbor-operator/pkg/rest"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	finalizerID = "robotaccount.finalizers.resource.goharbor.io"
)

var (
	ErrHarborCfgNotFound         = errors.New("harbor server configuration not found")
	ErrUnexpectedHarborCfgStatus = errors.New("status of Harbor server referred in configuration %s is unexpected")
)

// RobotAccountReconciler reconciles a RobotAccount object.
type Reconciler struct {
	*commonCtrl.Controller
	client.Client
	Scheme *runtime.Scheme
	Harbor *v2.Client
}

// New RobotAccount reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborCluster, nil, configStore)

	return r, nil
}

func (r *Reconciler) update(ctx context.Context, ra *goharborv1beta1.RobotAccount) error {
	if err := r.Client.Update(ctx, ra, &client.UpdateOptions{}); err != nil {
		return err
	}

	// Refresh object status to avoid problem
	namespacedName := types.NamespacedName{
		Name:      ra.Name,
		Namespace: ra.Namespace,
	}

	return r.Client.Get(ctx, namespacedName, ra)
}

// cleanRobotAccount remove the robot account from harbor.
func (r *Reconciler) cleanRobotAccount(ra *goharborv1beta1.RobotAccount) error {
	// get robot account
	_, err := r.Harbor.GetRobotAccountByID(ra.Status.ID)

	var robotAccountNotFound *robot.GetRobotByIDNotFound
	if err != nil && errors.As(err, &robotAccountNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	// delete robot account
	if err := r.Harbor.DeleteRobotAccount(ra.Status.ID); err != nil {
		return err
	}

	return nil
}

// createOrUpdateRobotAccount creates or updates a robot account.
func (r *Reconciler) createOrUpdateRobotAccount(ra *goharborv1beta1.RobotAccount) (*models.Robot, error) {
	// get robot account
	raGetIns, err := r.Harbor.GetRobotAccountByName(ra.Spec.Name)

	var robotAccountListNotFound *robot.ListRobotNotFound

	// fail to get robot account from harbor.
	if err != nil && !errors.As(err, &robotAccountListNotFound) {
		return nil, err
	}

	// robot account not found and create it.
	if err != nil && errors.As(err, &robotAccountListNotFound) {
		return r.Harbor.CreateRobotAccount(ra)
	}

	// err is nil and robot account is found
	// update robot account
	return r.Harbor.UpdateRobotAccount(raGetIns.ID, ra)
}

// setHarborClient sets up the Harbor client.
func (r *Reconciler) setHarbotClient(ctx context.Context, ra *goharborv1beta1.RobotAccount) error {
	harborCfg, err := r.getHarborServerConfig(ctx, ra.Spec.HarborServerConfig)
	if err != nil {
		return fmt.Errorf("error finding harborCfg: %w", err)
	}

	if harborCfg == nil {
		// Not exist
		return fmt.Errorf("%w: %s", ErrHarborCfgNotFound, ra.Spec.HarborServerConfig)
	}

	if harborCfg.Status.Status == goharborv1beta1.HarborServerConfigurationStatusUnknown || harborCfg.Status.Status == goharborv1beta1.HarborServerConfigurationStatusFail {
		return fmt.Errorf("%w harborCfg %s with %s", ErrUnexpectedHarborCfgStatus, harborCfg.Name, harborCfg.Status.Status)
	}

	// Create harbor client
	harborv2, err := harborClient.CreateHarborV2Client(ctx, r.Client, harborCfg)
	if err != nil {
		return err
	}

	r.Harbor = harborv2.WithContext(ctx)

	return nil
}

// delete removes the robot account from harbor.
func (r *Reconciler) delete(ctx context.Context, ra *goharborv1beta1.RobotAccount) error {
	if err := r.cleanRobotAccount(ra); err != nil {
		return err
	}

	if strings.ContainsString(ra.ObjectMeta.Finalizers, finalizerID) {
		// Execute and remove our finalizer from the finalizer list
		ra.ObjectMeta.Finalizers = strings.RemoveString(ra.ObjectMeta.Finalizers, finalizerID)
		if err := r.Client.Update(ctx, ra, &client.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
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
	ra := &goharborv1beta1.RobotAccount{}
	if err := r.Client.Get(ctx, req.NamespacedName, ra); err != nil {
		if apierr.IsNotFound(err) {
			// It could have been deleted after reconcile request coming in.
			log.Info(fmt.Sprintf("Harbor robotaccount %s does not exist", req.NamespacedName))

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get robotAccount %s error: %w", req.NamespacedName, err)
	}

	if err := r.setHarbotClient(ctx, ra); err != nil {
		log.Error(err, fmt.Sprintf("failed to set harbor client for robotAccount %s", req.NamespacedName))

		return ctrl.Result{}, fmt.Errorf("setHarbotClient error: %w", err)
	}

	// Check if the robot account is deleted, if so, delete robotaccount rc in k8s and robot account in harbor.
	if !ra.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.delete(ctx, ra)
	}

	// set Finalizer for the robot account rc.
	if !strings.ContainsString(ra.ObjectMeta.Finalizers, finalizerID) {
		ra.ObjectMeta.Finalizers = append(ra.ObjectMeta.Finalizers, finalizerID)
		if err := r.update(ctx, ra); err != nil {
			return ctrl.Result{}, err
		}
	}

	// create or update robot account.
	modelRobot, err := r.createOrUpdateRobotAccount(ra)
	if err != nil {
		log.Error(err, "failed to create robot account")

		return ctrl.Result{}, err
	}

	// Update status
	err = r.setStatus(ctx, ra, modelRobot.ID, "robot account updated", "", modelRobot.Secret)

	return ctrl.Result{}, err
}

// setStatus sets the status of the robot account.
func (r *Reconciler) setStatus(ctx context.Context, ra *goharborv1beta1.RobotAccount, id int64, reason, message, secret string) error {
	if id != 0 {
		ra.Status.ID = id
	}

	if secret != "" {
		ra.Status.Secret = secret
	}

	ra.Status.Reason = reason
	ra.Status.Message = message

	return r.Status().Update(ctx, ra)
}

func (r *Reconciler) getHarborServerConfig(ctx context.Context, name string) (*goharborv1beta1.HarborServerConfiguration, error) {
	hsc := &goharborv1beta1.HarborServerConfiguration{}
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
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1beta1.RobotAccount{}).
		Complete(r)
}

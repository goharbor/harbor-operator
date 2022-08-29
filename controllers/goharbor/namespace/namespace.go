package namespace

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	v2models "github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	harborClient "github.com/goharbor/harbor-operator/pkg/rest"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	"github.com/goharbor/harbor-operator/pkg/utils/consts"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	defaultSaName = "default"
	baseInt10     = 10
	baseBitSize   = 64
)

// New Namespace reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.Namespace, nil, configStore)

	return r, nil
}

// Reconciler reconciles a Namespace object.
type Reconciler struct {
	*commonCtrl.Controller
	client.Client
	Scheme *runtime.Scheme
	Harbor *v2.Client
}

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=goharbor.io,resources=pullsecretbindings,verbs=get;list;watch;create;delete

// Reconcile the Namespace.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) { //nolint:funlen
	log := r.Log.WithValues("namespace", req.NamespacedName)

	// Get the namespace object
	ns := &corev1.Namespace{}
	if err := r.Client.Get(ctx, req.NamespacedName, ns); err != nil {
		if apierr.IsNotFound(err) {
			// The resource may have been deleted after reconcile request coming in
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get namespace error: %w", err)
	}

	// Check if the ns is being deleted
	if !ns.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("namespace is being deleted", "name", ns.Name)

		return ctrl.Result{}, nil
	}

	// Get the binding list if existing
	bindings := &goharborv1.PullSecretBindingList{}
	if err := r.Client.List(ctx, bindings, &client.ListOptions{Namespace: req.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("list bindings error: %w", err)
	}

	// If auto is set for image rewrite rule
	harborCfg, err := r.findDefaultHarborCfg(ctx, log, ns)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error finding harborCfg: %w", err)
	}

	if harborCfg == nil {
		log.Info("no default hsc for namespace: ", req.Namespace, ", skip PSB creation")

		if err := r.removeStalePSB(ctx, log, bindings); err != nil {
			log.Info("error removing stale psb")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, nil
	}

	// Pull secret issuer is set and then check if the required default binding exists
	// Confirm the service account name
	// Use default SA if not set inside annotation
	saName := defaultSaName
	if setSa, ok := ns.Annotations[consts.AnnotationAccount]; ok {
		saName = setSa
	}
	// Check if custom service account exist
	sa := &corev1.ServiceAccount{}
	saNamespacedName := types.NamespacedName{
		Namespace: ns.Name,
		Name:      saName,
	}

	if err := r.Client.Get(ctx, saNamespacedName, sa); err != nil {
		if apierr.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("service account %s not found in namespace %s: %w", saName, ns.Name, err)
		}

		return ctrl.Result{}, fmt.Errorf("get service account %s in namespace %s error: %w", saName, ns.Name, err)
	}

	// Find PSB
	for _, bd := range bindings.Items {
		if bd.Spec.HarborServerConfig == harborCfg.Name && bd.Spec.ServiceAccount == saName {
			// Found it and reconcile is done
			// TODO: the PSB might be useless if the credentials or projects are changed
			// Need to check if the PSB is still valid.
			log.Info("psb exist for this namespace")

			return ctrl.Result{}, nil
		}
	}

	proj, projExist := ns.Annotations[consts.AnnotationProject]

	if !projExist || proj == "" {
		log.Info("annotation 'project' not set, skip reconciliation")

		return ctrl.Result{}, nil
	}

	if err := r.setHarborClient(ctx, log, harborCfg); err != nil {
		return ctrl.Result{}, err
	}

	var projName, projID, robotID string

	if projName, projID, robotID, err = r.validateHarborProjectAndRobot(log, ns); err != nil {
		return ctrl.Result{}, err
	}

	// PSB doesn't exist, create one
	log.Info("creating pull secret binding")

	psb, err := r.createPullSecretBinding(ctx, ns, harborCfg.Name, saName, robotID, projID)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("created pull secret binding", "name", psb.Name)

	// update namespace with updated annotation
	if proj == "*" {
		log.Info("update namespace annotations", "projectName", projName, "robotID", robotID)
		ns.Annotations[consts.AnnotationProject] = projName
		ns.Annotations[consts.AnnotationRobot] = robotID

		if err := r.Client.Update(ctx, ns, &client.UpdateOptions{}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *Reconciler) getNewBindingCR(ns string, harborCfg string, sa string) *goharborv1.PullSecretBinding {
	return &goharborv1.PullSecretBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.RandomName("Binding"),
			Namespace: ns,
		},
		Spec: goharborv1.PullSecretBindingSpec{
			HarborServerConfig: harborCfg,
			ServiceAccount:     sa,
		},
	}
}

func (r *Reconciler) validateProject(projectName string) (string, error) {
	var (
		proj *v2models.Project
		err  error
	)

	if proj, err = r.Harbor.GetProject(projectName); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", proj.ProjectID), nil
}

func (r *Reconciler) validateRobot(proj, robot string) error {
	if robot == "" {
		return errors.Errorf("robot should not be empty")
	}

	if proj == "" {
		return errors.Errorf("proj should not be empty")
	}

	robotID, err := strconv.ParseInt(robot, baseInt10, baseBitSize)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(proj, baseInt10, baseBitSize)
	if err != nil {
		return err
	}

	_, err = r.Harbor.GetRobotAccount(projectID, robotID)

	return err
}

func (r *Reconciler) createProjectAndRobot(proj string) (string, string, error) {
	projID, err := r.Harbor.EnsureProject(proj)
	if err != nil {
		return "", "", err
	}

	robot, err := r.Harbor.CreateRobotAccount(fmt.Sprintf("%d", projID))
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%d", projID), fmt.Sprintf("%d", robot.ID), nil
}

func (r *Reconciler) findDefaultHarborCfg(ctx context.Context, log logr.Logger, ns *corev1.Namespace) (*goharborv1.HarborServerConfiguration, error) {
	// check annotation first
	harborCfg, yes := ns.Annotations[consts.AnnotationHarborServer]
	if yes && harborCfg != "" {
		hsc := &goharborv1.HarborServerConfiguration{}

		err := r.Client.Get(ctx, types.NamespacedName{Name: harborCfg}, hsc)
		if err != nil {
			if apierr.IsNotFound(err) {
				log.Info("hsc specified in annotation doesn't exist")

				return nil, nil
			}

			return nil, fmt.Errorf("error when finding hsc specified in annotation: %w", err)
		}

		return hsc, nil
	}

	log.Info("no default hsc found in annotation for namespace " + ns.Name)

	// then find global default hsc
	hscs := &goharborv1.HarborServerConfigurationList{}

	err := r.Client.List(ctx, hscs)
	if err != nil {
		return nil, fmt.Errorf("error listing harborCfg: %w", err)
	}

	if len(hscs.Items) > 0 {
		for _, hsc := range hscs.Items {
			if hsc.Spec.Default {
				log.Info("found global default hsc: " + hsc.Name)

				return &hsc, nil
			}
		}
	}

	return nil, nil
}

func (r *Reconciler) removeStalePSB(ctx context.Context, log logr.Logger, bindings *goharborv1.PullSecretBindingList) error {
	if len(bindings.Items) > 0 {
		log.Info("removig stale psb in namespace")

		for i, bd := range bindings.Items {
			// Remove all the existing bindings as issuer is removed
			if err := r.Client.Delete(ctx, &bindings.Items[i], &client.DeleteOptions{}); err != nil {
				// Retry next time
				return fmt.Errorf("remove binding %s error: %w", bd.Name, err)
			}
		}
	}

	return nil
}

func (r *Reconciler) createPullSecretBinding(ctx context.Context, ns *corev1.Namespace, harborCfg, saName, robotID, projID string) (*goharborv1.PullSecretBinding, error) {
	defaultBinding := r.getNewBindingCR(ns.Name, harborCfg, saName)
	if err := controllerutil.SetControllerReference(ns, defaultBinding, r.Scheme); err != nil {
		return nil, fmt.Errorf("set ctrl reference error: %w", err)
	}

	defaultBinding.Spec.RobotID = robotID
	defaultBinding.Spec.ProjectID = projID

	if err := r.Client.Create(ctx, defaultBinding, &client.CreateOptions{}); err != nil {
		return nil, fmt.Errorf("create binding CR error: %w", err)
	}

	return defaultBinding, nil
}

func (r *Reconciler) setHarborClient(ctx context.Context, log logr.Logger, harborCfg *goharborv1.HarborServerConfiguration) error {
	// Create harbor client
	harborV2, err := harborClient.CreateHarborV2Client(ctx, r.Client, harborCfg)
	if err != nil {
		log.Error(err, "failed to create harbor client")

		return err
	}

	r.Harbor = harborV2.WithContext(ctx)

	return nil
}

func (r *Reconciler) validateHarborProjectAndRobot(log logr.Logger, ns *corev1.Namespace) (string, string, string, error) {
	var (
		err    error
		projID string
	)

	// Validate the annotation and create PSB is needed
	proj := ns.Annotations[consts.AnnotationProject]
	robotID, robotExist := ns.Annotations[consts.AnnotationRobot]

	if proj == "*" {
		log.Info("validate project and robot account")
		// Automatically generate project and robot account based on namespace name
		// TODO: should be more structure name since many clusters might share the same Harbor instance
		proj = strings.RandomName(ns.Name)
		projID, robotID, err = r.createProjectAndRobot(proj)

		if err != nil {
			log.Error(err, "Failed creating project and robot", "project", proj, "robot", robotID)

			return "", "", "", err
		}

		return proj, projID, robotID, nil
	}

	projID, err = r.validateProject(proj)
	if err != nil {
		log.Error(err, "Harbor annotation for project is invalid", "project", proj)

		return "", "", "", fmt.Errorf("project are invalid: %w", err)
	}

	if !robotExist {
		return "", "", "", errors.New("robotID is not set")
	}

	err = r.validateRobot(projID, robotID)
	if err != nil {
		log.Error(err, "annotation 'robotID'  is invalid", "robotID", robotID)

		return "", "", "", fmt.Errorf("robotID is invalid: %w", err)
	}

	return proj, projID, robotID, nil
}

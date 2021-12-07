package pullsecretbinding

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/registry/secret"
	"github.com/goharbor/harbor-operator/pkg/rest/model"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	"github.com/goharbor/harbor-operator/pkg/utils/consts"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
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
	defaultOwner = "harbor-operator-ns"
	regSecType   = "kubernetes.io/dockerconfigjson"
	datakey      = ".dockerconfigjson"
	finalizerID  = "psb.finalizers.resource.goharbor.io"
	defaultCycle = 5 * time.Minute
)

// New PullSecretBinding reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.PullSecretBinding, nil, configStore)

	return r, nil
}

// Reconciler reconciles a PullSecretBinding object
type Reconciler struct {
	*commonCtrl.Controller
	client.Client
	Scheme *runtime.Scheme
	Harbor *v2.Client
}

// +kubebuilder:rbac:groups=goharbor.io,resources=pullsecretbindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=pullsecretbindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=goharbor.io,resources=harborserverconfigurations,verbs=get
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;create;update
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;update;patch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, ferr error) {
	log := r.Log.WithValues("pullsecretbinding", req.NamespacedName)

	// Get the psb object
	bd := &goharborv1.PullSecretBinding{}
	if err := r.Client.Get(ctx, req.NamespacedName, bd); err != nil {
		if apierr.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get binding CR error: %w", err)
	}

	// Check binding resources
	server, sa, res, err := r.checkBindingRes(ctx, bd)
	if err != nil {
		return res, err
	} else {
		if server == nil || sa == nil {
			return res, err
		}
	}

	// Create harbor client
	harborv2, err := v2.NewWithServer(server)
	if err != nil {
		log.Error(err, "failed to create harbor client")

		return ctrl.Result{}, err
	}

	r.Harbor = harborv2.WithContext(ctx)

	// Check if the binding is being deleted
	if bd.ObjectMeta.DeletionTimestamp.IsZero() {
		if !strings.ContainsString(bd.ObjectMeta.Finalizers, finalizerID) {
			// Append finalizer
			bd.ObjectMeta.Finalizers = append(bd.ObjectMeta.Finalizers, finalizerID)
			if err := r.update(ctx, bd); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if strings.ContainsString(bd.ObjectMeta.Finalizers, finalizerID) {
			// Execute and remove our finalizer from the finalizer list
			if err := r.deleteExternalResources(bd); err != nil {
				return ctrl.Result{}, err
			}

			bd.ObjectMeta.Finalizers = strings.RemoveString(bd.ObjectMeta.Finalizers, finalizerID)
			if err := r.Client.Update(ctx, bd, &client.UpdateOptions{}); err != nil {
				return ctrl.Result{}, err
			}
		}

		log.Info("pull secret binding is being deleted")
		return ctrl.Result{}, nil
	}

	defer func() {
		if ferr != nil && bd.Status.Status != "error" {
			bd.Status.Status = "error"
			bd.Status.Message = ferr.Error()
			if err := r.Status().Update(ctx, bd, &client.UpdateOptions{}); err != nil {
				log.Error(err, "defer update status error", "cause", err)
			}
		}
	}()

	projID, robotID := parseIntID(bd.Spec.ProjectID), parseIntID(bd.Spec.RobotID)

	// Bind robot to service account
	// TODO: may cause dirty robots at the harbor project side
	// TODO: check secret binding by get secret and service account
	_, ok := bd.Annotations[consts.AnnotationRobotSecretRef]
	if !ok {
		// Need to create a new one as we only have one time to get the robot token
		robot, err := r.Harbor.GetRobotAccount(projID, robotID)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("create robot account error: %w", err)
		}

		// Make registry secret
		regsec, err := r.createRegSec(ctx, bd.Namespace, server.ServerURL, robot, bd)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("create registry secret error: %w", err)
		}
		// Add secret to service account
		if sa.ImagePullSecrets == nil {
			sa.ImagePullSecrets = make([]corev1.LocalObjectReference, 0)
		}
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: regsec.Name,
		})

		// Update
		if err := r.Client.Update(ctx, sa, &client.UpdateOptions{}); err != nil {
			return ctrl.Result{}, fmt.Errorf("update error: %w", err)
		}

		// Update binding
		if err := controllerutil.SetControllerReference(bd, regsec, r.Scheme); err != nil {
			r.Log.Error(err, "set controller reference", "owner", bd.ObjectMeta, "controlled", regsec.ObjectMeta)
		}
		setAnnotation(bd, consts.AnnotationRobotSecretRef, regsec.Name)
		if err := r.update(ctx, bd); err != nil {
			return ctrl.Result{}, fmt.Errorf("update error: %w", err)
		}
	}

	// TODO: add conditions
	if bd.Status.Status != "ready" {
		bd.Status.Status = "ready"
		if err := r.Status().Update(ctx, bd, &client.UpdateOptions{}); err != nil {
			if apierr.IsConflict(err) {
				log.Error(err, "failed to update status")
			} else {
				return ctrl.Result{}, err
			}
		}
	}

	// Loop
	return ctrl.Result{
		RequeueAfter: defaultCycle,
	}, nil
}

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1.PullSecretBinding{}).
		Complete(r)
}

func (r *Reconciler) update(ctx context.Context, binding *goharborv1.PullSecretBinding) error {
	if err := r.Client.Update(ctx, binding, &client.UpdateOptions{}); err != nil {
		return err
	}

	// Refresh object status to avoid problem
	namespacedName := types.NamespacedName{
		Name:      binding.Name,
		Namespace: binding.Namespace,
	}
	return r.Client.Get(ctx, namespacedName, binding)
}

func (r *Reconciler) getConfigData(ctx context.Context, hsc *goharborv1.HarborServerConfiguration) (*model.HarborServer, error) {
	s := &model.HarborServer{
		ServerURL: hsc.Spec.ServerURL,
		Insecure:  hsc.Spec.Insecure,
	}

	namespacedName := types.NamespacedName{
		Namespace: hsc.Spec.AccessCredential.Namespace,
		Name:      hsc.Spec.AccessCredential.AccessSecretRef,
	}
	sec := &corev1.Secret{}
	if err := r.Client.Get(ctx, namespacedName, sec); err != nil {
		return nil, fmt.Errorf("failed to get the configured secret with error: %w", err)
	}

	username, password, err := model.GetCredential(sec)
	if err != nil {
		return nil, fmt.Errorf("get credential error: %w", err)
	}

	s.Username = username
	s.Password = password

	return s, nil
}

func (r *Reconciler) checkBindingRes(ctx context.Context, psb *goharborv1.PullSecretBinding) (*model.HarborServer, *corev1.ServiceAccount, ctrl.Result, error) {
	// Get server configuration
	hsc, err := r.getHarborServerConfig(ctx, psb.Spec.HarborServerConfig)
	if err != nil {
		// Retry later
		return nil, nil, ctrl.Result{}, fmt.Errorf("get server configuration error: %w", err)
	}

	if hsc == nil {
		// Not exist
		r.Log.Info("harbor server configuration does not exists", "name", psb.Spec.HarborServerConfig)
		// Do not need to reconcile again
		return nil, nil, ctrl.Result{}, nil
	}

	if hsc.Status.Status == goharborv1.HarborServerConfigurationStatusUnknown || hsc.Status.Status == goharborv1.HarborServerConfigurationStatusFail {
		return nil, nil, ctrl.Result{}, fmt.Errorf("status of Harbor server referred in configuration %s is unexpected: %s", hsc.Name, hsc.Status.Status)
	}

	// Get the specified service account
	sa, err := r.getServiceAccount(ctx, psb.Namespace, psb.Spec.ServiceAccount)
	if err != nil {
		// Retry later
		return nil, nil, ctrl.Result{}, fmt.Errorf("get service account error: %w", err)
	}

	if sa == nil {
		// Not exist
		r.Log.Info("service account does not exist", "name", psb.Spec.ServiceAccount)
		// Do not need to reconcile again
		return nil, nil, ctrl.Result{}, nil
	}

	hs, err := r.getConfigData(ctx, hsc)
	if err != nil {
		return nil, nil, ctrl.Result{}, fmt.Errorf("get config data error: %w", err)
	}

	return hs, sa, ctrl.Result{}, nil
}

func (r *Reconciler) getHarborServerConfig(ctx context.Context, name string) (*goharborv1.HarborServerConfiguration, error) {
	hsc := &goharborv1.HarborServerConfiguration{}
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

func (r *Reconciler) getServiceAccount(ctx context.Context, ns, name string) (*corev1.ServiceAccount, error) {
	sc := &corev1.ServiceAccount{}
	namespacedName := types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}

	if err := r.Client.Get(ctx, namespacedName, sc); err != nil {
		if apierr.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return sc, nil
}

func (r *Reconciler) createRegSec(ctx context.Context, namespace string, registry string, robot *models.Robot, psb *goharborv1.PullSecretBinding) (*corev1.Secret, error) {
	auths := &secret.Object{
		Auths: map[string]*secret.Auth{},
	}
	auths.Auths[registry] = &secret.Auth{
		Username: robot.Name,
		Password: robot.Secret,
		Email:    fmt.Sprintf("%s@goharbor.io", robot.Name),
	}

	encoded := auths.Encode()

	regSec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.RandomName("regsecret"),
			Namespace: namespace,
			Annotations: map[string]string{
				consts.AnnotationSecOwner: defaultOwner,
			},
			OwnerReferences: []metav1.OwnerReference{{APIVersion: psb.APIVersion, Kind: psb.Kind, Name: psb.Name, UID: psb.UID}},
		},
		Type: regSecType,
		Data: map[string][]byte{
			datakey: encoded,
		},
	}

	return regSec, r.Client.Create(ctx, regSec, &client.CreateOptions{})
}

func (r *Reconciler) deleteExternalResources(bd *goharborv1.PullSecretBinding) error {
	if pro, ok := bd.Annotations[consts.AnnotationProject]; ok {
		if err := r.Harbor.DeleteProject(pro); err != nil {
			// TODO: handle delete error
			// Delete non-empty project will cause error?
			r.Log.Error(err, "delete external resources", "finalizer", finalizerID)
		}
	}

	return nil
}

func setAnnotation(obj *goharborv1.PullSecretBinding, key string, value string) {
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}

	obj.Annotations[key] = value
}

func parseIntID(id string) int64 {
	intID, _ := strconv.ParseInt(id, 10, 64)
	return intID
}

package configuration

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// New HarborConfiguration configmap reconciler.
func NewWithCm(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	return &CmReconciler{
		Log: ctrl.Log.WithName(application.GetName(ctx)).WithName("configuration-configmap-controller").WithValues("controller", "HarborConfiguration"),
	}, nil
}

const (
	// ConfigurationLabelKey is the key label for configuration.
	ConfigurationLabelKey = "goharbor.io/configuration"
)

// CmReconciler reconciles a configuration configmap.
type CmReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var configMapPredicate = predicate.Funcs{
	// configuration reconciler only watch create and update events, ignore
	// delete and generic events.
	CreateFunc: func(event event.CreateEvent) bool {
		return isConfiguration(event.Object)
	},
	UpdateFunc: func(event event.UpdateEvent) bool {
		return isConfiguration(event.ObjectNew)
	},
	DeleteFunc: func(event event.DeleteEvent) bool {
		return false
	},
	GenericFunc: func(event event.GenericEvent) bool {
		return false
	},
}

// isConfiguration checks whether the object has configuration anno.
func isConfiguration(obj metav1.Object) bool {
	if _, ok := obj.GetAnnotations()[ConfigurationLabelKey]; ok {
		return true
	}

	return false
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *CmReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(configMapPredicate).
		Complete(r)
}

func (r *CmReconciler) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	suffixes = append([]string{"CmConfiguration"}, suffixes...)

	return strings.NormalizeName(name, suffixes...)
}

// Reconcile does configuration reconcile.
func (r *CmReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("resource", req.NamespacedName)

	log.Info("Start reconciling")
	log.Info("Warning!!! ConfigMap configuration controller will be deprecated at v1.3, please use HarborConfiguration CRD.")

	defer func() {
		log.Info("Reconcile end", "result", res, "error", err)
	}()

	// get the configmap firstly
	cm := &corev1.ConfigMap{}
	if err = r.Client.Get(ctx, req.NamespacedName, cm); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			// Reconcile is done
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get harbor cluster configmap error: %w", err)
	}

	log.Info("Convert harbor configuration configmap to HarborConfiguration CR")

	oldConfig := []byte(cm.Data["config.yaml"])
	model := &goharborv1.HarborConfigurationModel{}

	if err = yaml.Unmarshal(oldConfig, model); err != nil {
		return ctrl.Result{}, fmt.Errorf("error unmarshal configmap configuration to HarborConfigurationSpec: %w", err)
	}

	hc := &goharborv1.HarborConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.Name,
			Namespace: cm.Namespace,
		},
		Spec: goharborv1.HarborConfigurationSpec{
			HarborClusterRef: cm.GetAnnotations()[ConfigurationLabelKey],
			Configuration:    *model,
		},
	}

	if err = r.createOrUpdateHarborConfiguration(ctx, hc); err != nil {
		return ctrl.Result{}, fmt.Errorf("error create or update harbor configuration: %w", err)
	}

	// delete configmap
	if err = r.Client.Delete(ctx, cm); err != nil {
		return ctrl.Result{}, fmt.Errorf("error delete configmap: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *CmReconciler) createOrUpdateHarborConfiguration(ctx context.Context, hc *goharborv1.HarborConfiguration) error {
	old := &goharborv1.HarborConfiguration{}

	err := r.Client.Get(ctx, types.NamespacedName{Namespace: hc.Namespace, Name: hc.Name}, old)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// create hc
			r.Log.Info("Create HarborConfiguration", "hc", hc)

			return r.Client.Create(ctx, hc)
		}
	}

	// if hc exist, update it
	new := old.DeepCopy()
	new.Spec = hc.Spec

	r.Log.Info("Update HarborConfiguration", "hc", new)

	return r.Client.Update(ctx, new)
}

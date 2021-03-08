package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/harbor"
	"github.com/ovh/configstore"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"
)

// pwdFields deines configuration password related fidlds list.
var pwdFields = []string{"email_password", "ldap_search_password", "uaa_client_secret", "oidc_client_secret"}

// New HarborConfiguration reconciler.
func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	return &Reconciler{
		Log: ctrl.Log.WithName(application.GetName(ctx)).WithName("configuration-controller").WithValues("controller", name),
	}, nil
}

const (
	// ConfigurationLabelKey is the key label for configuration.
	ConfigurationLabelKey = "goharbor.io/configuration"
	// ConfigurationApplyError is the reason of condition.
	ConfigurationApplyError = "ConfigurationApplyError"
	// ConfigurationApplySuccess is the reason of condition.
	ConfigurationApplySuccess = "ConfigurationApplySuccess"
)

// Reconciler reconciles a configuration configmap.
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var configMapPredicate = predicate.Funcs{
	// configuration reconciler only watch create and update events, ignore
	// delete and generic events.
	CreateFunc: func(event event.CreateEvent) bool {
		return isConfiguration(event.Meta)
	},
	UpdateFunc: func(event event.UpdateEvent) bool {
		return isConfiguration(event.MetaNew)
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

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(configMapPredicate).
		Complete(r)
}

// Reconcile does configuration reconcile.
func (r *Reconciler) Reconcile(req ctrl.Request) (res ctrl.Result, err error) {
	ctx := context.TODO()
	log := r.Log.WithValues("resource", req.NamespacedName)

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
	// get harborcluster name from annotataions
	harborClusterName := cm.GetAnnotations()[ConfigurationLabelKey]
	if len(harborClusterName) == 0 {
		// if configmap value is invalid, not do reconcile
		return ctrl.Result{}, nil
	}
	// get harbor cluster
	cluster := &v1alpha2.HarborCluster{}
	if err = r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: harborClusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			// Reconcile is done
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get harbor cluster cr error: %w", err)
	}

	log.Info("Get configmap and harbor cluster cr successfully", "configmap", cm, "harborcluster", cluster)

	defer func() {
		log.Info("Reconcile end", "result", res, "error", err, "updateStatusErr", r.UpdateStatus(ctx, err, cluster))
	}()

	harborClient, err := r.getHarborClient(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get harbor client error: %w", err)
	}
	// assemble config payload
	jsonPayload, err := r.assembleConfig(ctx, cm)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("assemble configuration error: %w", err)
	}

	// apply configuration
	if err = harborClient.ApplyConfiguration(ctx, jsonPayload); err != nil {
		return ctrl.Result{}, fmt.Errorf("apply harbor configuration error: %w", err)
	}

	log.Info("Apply harbor configuration successfully", "configmap", cm, "harborcluster", cluster.Name)

	return ctrl.Result{}, nil
}

// assembleConfig assembles password filed from secret.
func (r *Reconciler) assembleConfig(ctx context.Context, cm *corev1.ConfigMap) (jsonPayload []byte, err error) {
	// configuration payload
	payload := cm.Data["config.yaml"]
	config := make(map[string]interface{})

	if err = yaml.Unmarshal([]byte(payload), &config); err != nil {
		return nil, fmt.Errorf("unmarshal config payload error: %w", err)
	}

	isPwdField := func(field string) bool {
		for _, v := range pwdFields {
			if field == v {
				return true
			}
		}

		return false
	}

	for itemKey, itemValue := range config {
		if isPwdField(itemKey) {
			// password field, read password from secret.
			secret := &corev1.Secret{}
			// itemValue is secret name.
			secretName, ok := itemValue.(string)
			if !ok {
				return nil, fmt.Errorf("config field %s's value %v, type is invalid, should be string", itemKey, itemValue)
			}
			// get secret.
			if err = r.Client.Get(ctx, types.NamespacedName{Namespace: cm.Namespace, Name: secretName}, secret); err != nil {
				return nil, fmt.Errorf("get config field %s value from secret %s error: %w", itemKey, itemValue, err)
			}
			// itemKey is the secret data key.
			config[itemKey] = secret.Data[itemKey]
		}
	}

	return json.Marshal(config)
}

// getHarborClient gets harbor client.
func (r *Reconciler) getHarborClient(ctx context.Context, cluster *v1alpha2.HarborCluster) (harbor.Client, error) {
	if cluster == nil {
		return nil, fmt.Errorf("harbor cluster can not be nil")
	}

	url := cluster.Spec.ExternalURL
	if len(url) == 0 {
		return nil, fmt.Errorf("harbor url is invalid")
	}

	var opts []harbor.ClientOption

	adminSecretRef := cluster.Spec.HarborAdminPasswordRef
	if len(adminSecretRef) > 0 {
		// fetch admin password
		secret := &corev1.Secret{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: adminSecretRef}, secret); err != nil {
			return nil, fmt.Errorf("get harbor admin secret error: %w", err)
		}

		password := string(secret.Data["secret"])
		opts = append(opts, harbor.WithCredential("admin", password))
	}

	return harbor.NewClient(url, opts...), nil
}

// UpdateStatus updates harbor cluster status.
func (r *Reconciler) UpdateStatus(ctx context.Context, err error, cluster *v1alpha2.HarborCluster) error {
	now := metav1.Now()
	cond := v1alpha2.HarborClusterCondition{
		Type:               v1alpha2.ConfigurationReady,
		LastTransitionTime: &now,
	}

	if err != nil {
		cond.Status = corev1.ConditionFalse
		cond.Reason = ConfigurationApplyError
		cond.Message = err.Error()
	} else {
		cond.Status = corev1.ConditionTrue
		cond.Reason = ConfigurationApplySuccess
		cond.Message = "harbor configuraion has been applied successfully."
	}

	var found bool

	for i, c := range cluster.Status.Conditions {
		if c.Type == cond.Type {
			found = true

			if c.Status != cond.Status ||
				c.Reason != cond.Reason ||
				c.Message != cond.Message {
				cluster.Status.Conditions[i] = cond
			}

			break
		}
	}

	if !found {
		cluster.Status.Conditions = append(cluster.Status.Conditions, cond)
	}
	// update rivision
	cluster.Status.Revision = time.Now().UnixNano()

	return r.Client.Status().Update(ctx, cluster)
}

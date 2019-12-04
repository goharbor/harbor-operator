package harbor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/api/extensions/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
	"github.com/ovh/harbor-operator/pkg/factories/logger"
)

// Reconciler reconciles a Harbor object
type Reconciler struct {
	client.Client

	Name    string
	Version string

	Log    logr.Logger
	Scheme *runtime.Scheme

	RestConfig *rest.Config
}

func (r *Reconciler) GetVersion() string {
	return r.Version
}

func (r *Reconciler) GetName() string {
	return r.Name
}

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=harbors,verbs=get;list;watch
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=harbors/status,verbs=get;update;patch

// nolint:funlen,gocognit
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()
	application.SetName(&ctx, r.GetName())
	application.SetVersion(&ctx, r.GetVersion())

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"Harbor.Namespace": req.Namespace,
		"Harbor.Name":      req.Name,
	})
	defer span.Finish()

	span.LogFields(
		log.String("Harbor.Namespace", req.Namespace),
		log.String("Harbor.Name", req.Name),
	)

	reqLogger := r.Log.WithValues("Request", req.NamespacedName, "Harbor.Namespace", req.Namespace, "Harbor.Name", req.Name)

	logger.Set(&ctx, reqLogger)

	// Fetch the Harbor instance
	harbor := &containerregistryv1alpha1.Harbor{}

	err := r.Client.Get(ctx, req.NamespacedName, harbor)
	if err != nil {
		if apierrs.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			reqLogger.Info("Harbor does not exists")
			return reconcile.Result{}, nil
		}

		// Error reading the object
		return reconcile.Result{}, err
	}

	result := reconcile.Result{}

	if !harbor.ObjectMeta.DeletionTimestamp.IsZero() {
		reqLogger.Info("harbor is being deleted")
		return result, nil
	}

	// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#status-subresource
	defer func() {
		err := r.Status().Update(ctx, harbor)
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot update status")
		}
	}()

	// TODO do it asynchronously but do not
	// forget to wait for completion before return
	health, err := r.GetHealth(ctx, harbor)
	if err != nil {
		result.Requeue = true

		err = r.UpdateCondition(ctx, harbor, containerregistryv1alpha1.ReadyConditionType, corev1.ConditionFalse, errors.Cause(err).Error(), err.Error())
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot set condition", "Harbor.Status.Condition.Type", containerregistryv1alpha1.ReadyConditionType, "Harbor.Status.Condition.Value", corev1.ConditionFalse)
		}
	} else {
		if health.IsHealthy() {
			err = r.UpdateCondition(ctx, harbor, containerregistryv1alpha1.ReadyConditionType, corev1.ConditionTrue)
			if err != nil {
				result.Requeue = true
				reqLogger.Error(err, "cannot set condition", "Harbor.Status.Condition.Type", containerregistryv1alpha1.ReadyConditionType, "Harbor.Status.Condition.Value", corev1.ConditionTrue)
			}
		} else {
			// Hide error, just try again later
			reqLogger.Info("not ready yet, trying again later")

			result.RequeueAfter = 2 * time.Second

			err = r.UpdateCondition(ctx, harbor, containerregistryv1alpha1.ReadyConditionType, corev1.ConditionFalse, "harbor-component", fmt.Sprintf("at least an Harbor component failed: %+v", health.GetUnhealthyComponents()))
			if err != nil {
				result.Requeue = true
				reqLogger.Error(err, "cannot set condition", "Harbor.Status.Condition.Type", containerregistryv1alpha1.ReadyConditionType, "Harbor.Status.Condition.Value", corev1.ConditionTrue)
			}
		}
	}

	if harbor.Status.ObservedGeneration != harbor.ObjectMeta.Generation {
		harbor.Status.ObservedGeneration = harbor.ObjectMeta.Generation

		err := r.UpdateCondition(ctx, harbor, containerregistryv1alpha1.AppliedConditionType, corev1.ConditionFalse, "new", "new generation detected")
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot set condition", "Harbor.Status.Condition.Type", containerregistryv1alpha1.AppliedConditionType, "Harbor.Status.Condition.Value", corev1.ConditionFalse)
		}
	}

	switch r.GetConditionStatus(ctx, harbor, containerregistryv1alpha1.AppliedConditionType) {
	case corev1.ConditionTrue: // Already applied
		err := r.Create(ctx, harbor)
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot create")
		}
	default: // apply failed
		err := r.UpdateCondition(ctx, harbor, containerregistryv1alpha1.AppliedConditionType, corev1.ConditionFalse)
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot set condition", "Harbor.Status.Condition.Type", containerregistryv1alpha1.AppliedConditionType, "Harbor.Status.Condition.Value", corev1.ConditionFalse)
		}

		err = r.Apply(ctx, harbor)
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot apply")

			break
		}

		err = r.UpdateCondition(ctx, harbor, containerregistryv1alpha1.AppliedConditionType, corev1.ConditionTrue)
		if err != nil {
			result.Requeue = true

			reqLogger.Error(err, "cannot set condition", "Harbor.Status.Condition.Type", containerregistryv1alpha1.AppliedConditionType, "Harbor.Status.Condition.Value", corev1.ConditionTrue)
		}
	}

	return result, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&containerregistryv1alpha1.Harbor{}).
		Owns(&appsv1.Deployment{}).
		Owns(&certv1.Certificate{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&extv1.Ingress{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

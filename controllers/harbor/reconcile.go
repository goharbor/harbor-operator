package harbor

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

// +kubebuilder:rbac:groups=goharbor.io,resources=harbors,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=harbors/status,verbs=get;update;patch

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
	harbor := &goharborv1alpha1.Harbor{}

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

	var g errgroup.Group

	g.Go(func() error {
		err = r.UpdateReadyStatus(ctx, &result, harbor)
		return errors.Wrapf(err, "type=%s", goharborv1alpha1.ReadyConditionType)
	})

	g.Go(func() error {
		err = r.UpdateAppliedStatus(ctx, &result, harbor)
		return errors.Wrapf(err, "type=%s", goharborv1alpha1.AppliedConditionType)
	})

	err = g.Wait()
	if err != nil {
		return result, errors.Wrap(err, "cannot set status")
	}

	return result, r.UpdateStatus(ctx, &result, harbor)
}

func (r *Reconciler) UpdateAppliedStatus(ctx context.Context, result *ctrl.Result, harbor *goharborv1alpha1.Harbor) error {
	if harbor.Status.ObservedGeneration != harbor.ObjectMeta.Generation {
		harbor.Status.ObservedGeneration = harbor.ObjectMeta.Generation

		err := r.UpdateCondition(ctx, harbor, goharborv1alpha1.AppliedConditionType, corev1.ConditionFalse, "new", "new generation detected")
		if err != nil {
			result.Requeue = true

			return errors.Wrapf(err, "value=%s", corev1.ConditionFalse)
		}
	}

	switch r.GetConditionStatus(ctx, harbor, goharborv1alpha1.AppliedConditionType) {
	case corev1.ConditionTrue: // Already applied
		// Anyway, reconciler is triggered, so at least one child resource has been deleted
		// Try to recreate children
		err := r.Create(ctx, harbor)
		if err != nil {
			result.Requeue = true

			err := r.UpdateCondition(ctx, harbor, goharborv1alpha1.AppliedConditionType, corev1.ConditionFalse, err.Error())
			if err != nil {
				result.Requeue = true

				return errors.Wrapf(err, "value=%s", corev1.ConditionFalse)
			}

			return nil
		}
	default: // Not yet applied
		err := r.UpdateCondition(ctx, harbor, goharborv1alpha1.AppliedConditionType, corev1.ConditionFalse)
		if err != nil {
			result.Requeue = true

			return errors.Wrapf(err, "value=%s", corev1.ConditionFalse)
		}

		err = r.Apply(ctx, harbor)
		if err != nil {
			err := r.UpdateCondition(ctx, harbor, goharborv1alpha1.AppliedConditionType, corev1.ConditionFalse, err.Error())
			if err != nil {
				result.Requeue = true

				return errors.Wrapf(err, "value=%s", corev1.ConditionFalse)
			}

			return nil
		}

		err = r.UpdateCondition(ctx, harbor, goharborv1alpha1.AppliedConditionType, corev1.ConditionTrue)
		if err != nil {
			result.Requeue = true

			return errors.Wrapf(err, "value=%s", corev1.ConditionTrue)
		}
	}

	return nil
}

func (r *Reconciler) UpdateReadyStatus(ctx context.Context, result *ctrl.Result, harbor *goharborv1alpha1.Harbor) error {
	// TODO do it asynchronously but do not
	// forget to wait for completion before return
	health, err := r.HealthClient.GetByProxy(ctx, harbor)
	if err != nil {
		result.Requeue = true

		err = r.UpdateCondition(ctx, harbor, goharborv1alpha1.ReadyConditionType, corev1.ConditionFalse, errors.Cause(err).Error(), err.Error())
		if err != nil {
			result.Requeue = true

			return errors.Wrapf(err, "value=%s", corev1.ConditionFalse)
		}
	} else {
		if health.IsHealthy() {
			err = r.UpdateCondition(ctx, harbor, goharborv1alpha1.ReadyConditionType, corev1.ConditionTrue)
			if err != nil {
				result.Requeue = true

				return errors.Wrapf(err, "value=%s", corev1.ConditionTrue)
			}
		} else {
			// Hide error, just try again later
			logger.Get(ctx).Info("not ready yet, trying again later")

			result.RequeueAfter = DefaultRequeueWait

			err = r.UpdateCondition(ctx, harbor, goharborv1alpha1.ReadyConditionType, corev1.ConditionFalse, "harbor-component", fmt.Sprintf("at least an Harbor component failed: %+v", health.GetUnhealthyComponents()))
			if err != nil {
				result.Requeue = true

				return errors.Wrapf(err, "value=%s", corev1.ConditionTrue)
			}
		}
	}

	return nil
}

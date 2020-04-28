package common

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"

	serrors "github.com/goharbor/harbor-operator/pkg/controllers/common/errors"
	sgraph "github.com/goharbor/harbor-operator/pkg/controllers/common/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	sstatus "github.com/goharbor/harbor-operator/pkg/status"
)

type Controller struct {
	client.Client

	Name    string
	Version string

	Scheme *runtime.Scheme
}

func NewController(name, version string) *Controller {
	return &Controller{
		Name:    name,
		Version: version,
	}
}

func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	c.Client = mgr.GetClient()
	c.Scheme = mgr.GetScheme()

	return nil
}

func (c *Controller) GetVersion() string {
	return c.Version
}

func (c *Controller) GetName() string {
	return c.Name
}

func (c *Controller) GetAndFilter(ctx context.Context, key client.ObjectKey, obj runtime.Object) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getAndFilter", opentracing.Tags{})
	defer span.Finish()

	err := c.Client.Get(ctx, key, obj)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *Controller) Reconcile(ctx context.Context, resource resources.Resource) (ctrl.Result, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commonReconcile", opentracing.Tags{
		"Resource": resource,
	})
	defer span.Finish()

	if !resource.GetDeletionTimestamp().IsZero() {
		logger.Get(ctx).Info("Object is being deleted")
		return ctrl.Result{}, nil
	}

	owner.Set(&ctx, resource)

	err := c.Run(ctx, resource)

	return c.HandleError(ctx, resource, err)
}

func (c *Controller) applyAndCheck(ctx context.Context, node graph.Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "applyAndCheck", opentracing.Tags{})
	defer span.Finish()

	err := c.Apply(ctx, node)
	if err != nil {
		return errors.Wrap(err, "apply")
	}

	err = c.EnsureReady(ctx, node)

	return errors.Wrap(err, "ready")
}

func (c *Controller) preUpdateData(ctx context.Context, u *unstructured.Unstructured) (bool, error) {
	err := status.Augment(u)
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to augment resource status")
	}

	data := u.UnstructuredContent()
	defer u.SetUnstructuredContent(data)

	generation := u.GetGeneration()

	observedGeneration, found, err := unstructured.NestedInt64(data, "status", "observedGeneration")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get observed generation")
	}

	conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get conditions")
	}

	// New generation
	if !found || generation == observedGeneration {
		err := unstructured.SetNestedField(data, generation, "status", "observedGeneration")
		if err != nil {
			return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to set observed generation")
		}

		conditions, err := sstatus.UpdateCondition(ctx, []interface{}{}, status.ConditionInProgress, corev1.ConditionTrue, "newGeneration", "New generation detected")
		if err != nil {
			return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to update condition")
		}

		err = unstructured.SetNestedSlice(data, conditions, "status", "conditions")
		if err != nil {
			return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to update condition")
		}

		return false, nil
	}

	s, err := sstatus.GetConditionStatus(ctx, conditions, status.ConditionInProgress)
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, fmt.Sprintf("unable to check %s condition", status.ConditionInProgress))
	}

	if s == corev1.ConditionFalse {
		// TODO Check what triggered the event
		return true, nil
	}

	return false, nil
}

func (c *Controller) Run(ctx context.Context, owner runtime.Object) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commonRun", opentracing.Tags{})
	defer span.Finish()

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(owner)
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to convert resource to unstuctured")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	stop, err := c.preUpdateData(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot update observedGeneration")
	}

	if stop {
		logger.Get(ctx).Info("nothing to do")
		return nil
	}

	err = c.Client.Status().Update(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot update status")
	}

	return sgraph.Get(ctx).Run(ctx, c.applyAndCheck)
}

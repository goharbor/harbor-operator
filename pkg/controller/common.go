package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/kustomize/kstatus/status"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
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

	ConfigStore *configstore.Store
	rm          ResourceManager
	Log         logr.Logger
	Scheme      *runtime.Scheme
}

func NewController(ctx context.Context, name string, rm ResourceManager, config *configstore.Store) *Controller {
	return &Controller{
		Name:        name,
		Version:     application.GetVersion(ctx),
		rm:          rm,
		Log:         ctrl.Log.WithName("controller").WithName(name),
		ConfigStore: config,
	}
}

func (c *Controller) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
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

func (c *Controller) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"Resource.Namespace": req.Namespace,
		"Resource.Name":      req.Name,
		"Controller.Name":    c.GetName(),
	})
	defer span.Finish()

	span.LogFields(
		log.String("Resource.Namespace", req.Namespace),
		log.String("Resource.Name", req.Name),
	)

	logger.Set(&ctx, c.Log)
	ctx = c.PopulateContext(ctx, req)
	l := logger.Get(ctx)

	// Fetch the Registry instance
	object := c.rm.NewEmpty(ctx)

	ok, err := c.GetAndFilter(ctx, req.NamespacedName, object)
	if err != nil {
		// Error reading the object
		return reconcile.Result{}, err
	}

	if !ok {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		l.Info("Object does not exists")
		return reconcile.Result{}, nil
	}

	if !object.GetDeletionTimestamp().IsZero() {
		logger.Get(ctx).Info("Object is being deleted")
		return ctrl.Result{}, nil
	}

	owner.Set(&ctx, object)

	err = c.Run(ctx, object)

	return c.HandleError(ctx, object, err)
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

	generation := u.GetGeneration()

	observedGeneration, found, err := unstructured.NestedInt64(data, "status", "observedGeneration")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get observed generation")
	}

	conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get conditions")
	}

	ok, err := c.preUpdateGenerationData(ctx, found, observedGeneration, generation, conditions, data)
	if err != nil {
		return false, err
	}

	u.SetUnstructuredContent(data)

	return ok, nil
}

func (c *Controller) preUpdateGenerationData(ctx context.Context, found bool, observedGeneration, generation int64, conditions []interface{}, data map[string]interface{}) (bool, error) {
	// New generation
	if !found || generation != observedGeneration {
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

	// TODO Check what triggered the event

	return s == corev1.ConditionFalse, nil
}

func (c *Controller) Run(ctx context.Context, owner resources.Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{})
	defer span.Finish()

	err := c.rm.AddResources(ctx, owner)
	if err != nil {
		return errors.Wrap(err, "cannot add resources")
	}

	err = c.prepareStatus(ctx, owner)
	if err != nil {
		return errors.Wrap(err, "cannot prepare owner status")
	}

	return sgraph.Get(ctx).Run(ctx)
}

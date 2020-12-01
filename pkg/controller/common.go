package controller

import (
	"context"

	"github.com/go-logr/logr"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/opentracing/opentracing-go"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceManager interface {
	AddResources(context.Context, resources.Resource) error
	NewEmpty(context.Context) resources.Resource
}

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
		Log:         ctrl.Log.WithName(application.GetName(ctx)).WithName("controller").WithValues("controller", name),
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "getAndFilter")
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
	ctx := c.NewContext(req)

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"resource.namespace": req.Namespace,
		"resource.name":      req.Name,
		"controller":         c.GetName(),
	})
	defer span.Finish()

	l := logger.Get(ctx)

	// Fetch the instance

	object := c.rm.NewEmpty(ctx)

	ok, err := c.GetAndFilter(ctx, req.NamespacedName, object)
	if err != nil {
		// Error reading the object
		return ctrl.Result{}, err
	}

	if !ok {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		l.Info("Object does not exists")

		return ctrl.Result{}, nil
	}

	if !object.GetDeletionTimestamp().IsZero() {
		logger.Get(ctx).Info("Object is being deleted")

		return ctrl.Result{}, nil
	}

	owner.Set(&ctx, object)

	err = c.Run(ctx, object)

	if err != nil {
		return c.HandleError(ctx, object, err)
	}

	return ctrl.Result{}, c.SetSuccessStatus(ctx, object)
}

func (c *Controller) applyAndCheck(ctx context.Context, node graph.Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "applyAndCheck")
	defer span.Finish()

	res, ok := node.(*Resource)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", node), serrors.OperatorReason, "unable to apply resource")
	}

	err := c.Apply(ctx, res)
	if err != nil {
		return errors.Wrap(err, "apply")
	}

	err = c.EnsureReady(ctx, res)

	return errors.Wrap(err, "check")
}

func (c *Controller) Run(ctx context.Context, owner resources.Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "run")
	defer span.Finish()

	logger.Get(ctx).V(1).Info("Reconciling object")

	if err := c.rm.AddResources(ctx, owner); err != nil {
		return errors.Wrap(err, "cannot add resources")
	}

	if err := c.prepareStatus(ctx, owner); err != nil {
		return errors.Wrap(err, "cannot prepare owner status")
	}

	return sgraph.Get(ctx).Run(ctx)
}

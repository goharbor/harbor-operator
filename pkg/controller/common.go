package controller

import (
	"context"

	"github.com/go-logr/logr"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceManager interface {
	AddResources(context.Context, resources.Resource) error
	NewEmpty(context.Context) resources.Resource
}

type Controller struct {
	client.Client

	BaseController controllers.Controller
	Version        string
	GitCommit      string

	deletableResources map[schema.GroupVersionKind]struct{}

	ConfigStore     *configstore.Store
	rm              ResourceManager
	Log             logr.Logger
	Scheme          *runtime.Scheme
	DiscoveryClient *discovery.DiscoveryClient
}

func NewController(ctx context.Context, base controllers.Controller, rm ResourceManager, config *configstore.Store) *Controller {
	version := application.GetVersion(ctx)
	gitCommit := application.GetGitCommit(ctx)

	logValues := []interface{}{
		"controller", base.String(),
		"version", version,
		"git.commit", gitCommit,
	}

	return &Controller{
		BaseController:     base,
		Version:            application.GetVersion(ctx),
		GitCommit:          gitCommit,
		rm:                 rm,
		Log:                ctrl.Log.WithName(application.GetName(ctx)).WithName("controller").WithValues(logValues...),
		deletableResources: application.GetDeletableResources(ctx),
		ConfigStore:        config,
	}
}

func (c *Controller) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	c.Client = mgr.GetClient()
	c.Scheme = mgr.GetScheme()
	c.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(mgr.GetConfig())

	return nil
}

func (c *Controller) GetGitCommit() string {
	return c.GitCommit
}

func (c *Controller) GetVersion() string {
	return c.Version
}

func (c *Controller) GetName() string {
	return c.BaseController.String()
}

func (c *Controller) GetAndFilter(ctx context.Context, key client.ObjectKey, obj client.Object) (bool, error) {
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

func (c *Controller) AreNetworkPoliciesEnabled(ctx context.Context, resource resources.Resource) (bool, error) {
	for name, value := range resource.GetAnnotations() {
		if name == harbormetav1.NetworkPoliciesAnnotationName {
			return value == harbormetav1.NetworkPoliciesAnnotationEnabled, nil
		}
	}

	networkPoliciesEnabled, err := config.GetBool(c.ConfigStore, config.NetworkPoliciesEnabledKey, config.DefaultNetworkPoliciesEnabled)

	return networkPoliciesEnabled, errors.Wrap(err, "get boolean config")
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = c.PopulateContext(ctx, req)

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

	if err := c.Run(ctx, object); err != nil {
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

	if err := c.PrepareStatus(ctx, owner); err != nil {
		return errors.Wrap(err, "cannot prepare owner status")
	}

	if err := c.Mark(ctx, owner); err != nil {
		return errors.Wrap(err, "cannot mark resources")
	}

	return sgraph.Get(ctx).Run(ctx)
}

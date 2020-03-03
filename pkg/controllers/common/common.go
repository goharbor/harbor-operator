package common

import (
	"context"

	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

type Controller struct {
	client.Client

	Name    string
	Version string

	Scheme *runtime.Scheme

	graph graph.Manager
}

func NewController(name, version string) *Controller {
	return &Controller{
		Name:    name,
		Version: version,
		graph:   graph.NewResourceManager(),
	}
}

func (r *Controller) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return nil
}

func (r *Controller) GetVersion() string {
	return r.Version
}

func (r *Controller) GetName() string {
	return r.Name
}

func (r *Controller) GetAndFilter(ctx context.Context, key client.ObjectKey, obj runtime.Object) (bool, error) {
	err := r.Client.Get(ctx, key, obj)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *Controller) Reconcile(ctx context.Context, resource resources.Resource) (ctrl.Result, error) {
	if !resource.GetDeletionTimestamp().IsZero() {
		logger.Get(ctx).Info("Object is being deleted")
		return ctrl.Result{}, nil
	}

	owner.Set(&ctx, resource)

	err := c.Run(ctx, resource)
	return c.HandleError(err)
}

func (c *Controller) applyAndCheck(ctx context.Context, node graph.Resource) error {
	err := c.Apply(ctx, node)
	if err != nil {
		return errors.Wrap(err, "apply")
	}

	err = c.EnsureReady(ctx, node)
	return errors.Wrap(err, "ready")
}

func (c *Controller) Run(ctx context.Context, owner runtime.Object) error {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(owner)
	if err != nil {
		logger.Get(ctx).Error(err, "Unable to convert resource to unstuctured")
		return nil
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	err = status.Augment(u)
	if err != nil {
		logger.Get(ctx).Error(err, "Unable to augment resource")
		return nil
	}

	return c.graph.Run(ctx, c.applyAndCheck)
}

package controller

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/resources/mutation"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
)

type ResourceManager interface {
	AddResources(context.Context, resources.Resource) error
	NewEmpty(context.Context) resources.Resource
}

type Resource struct {
	mutable   resources.Mutable
	checkable resources.Checkable
	resource  resources.Resource
}

func (c *Controller) AddUnsctructuredToManage(ctx context.Context, resource *unstructured.Unstructured, dependencies ...graph.Resource) (graph.Resource, error) { // nolint:interfacer
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewUnstructured(c.GlobalMutateFn(ctx)),
		checkable: statuscheck.UnstructuredCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

func (c *Controller) AddServiceToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewService(c.GlobalMutateFn(ctx)),
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

func (c *Controller) AddCertificateToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewCertificate(c.GlobalMutateFn(ctx)),
		checkable: statuscheck.CertificateCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

func (c *Controller) AddIngressToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewIngress(c.GlobalMutateFn(ctx)),
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

func (c *Controller) AddSecretToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewSecret(c.GlobalMutateFn(ctx)),
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

func (c *Controller) AddConfigMapToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewConfigMap(c.GlobalMutateFn(ctx)),
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

func (c *Controller) AddDeploymentToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   mutation.NewDeployment(c.DeploymentMutateFn(ctx, dependencies...)),
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.applyAndCheck)
}

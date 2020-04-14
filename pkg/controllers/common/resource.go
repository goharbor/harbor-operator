package common

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	sgraph "github.com/goharbor/harbor-operator/pkg/controllers/common/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
)

type Resource struct {
	mutable   resources.Mutable
	checkable resources.Checkable
	resource  resources.Resource
}

func (c *Controller) AddInstantResourceToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   c.GlobalMutateFn(ctx),
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(res, dependencies)
}

func (c *Controller) AddUnsctructuredToManage(ctx context.Context, resource *unstructured.Unstructured, dependencies ...graph.Resource) (graph.Resource, error) { // nolint:interfacer
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   c.GlobalMutateFn(ctx),
		checkable: statuscheck.UnstructuredCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(res, dependencies)
}

func (c *Controller) AddBasicObjectToManage(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   c.GlobalMutateFn(ctx),
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(res, dependencies)
}

func (c *Controller) AddDeploymentToManage(ctx context.Context, resource *appsv1.Deployment, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   c.DeploymentMutateFn(ctx),
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(res, dependencies)
}

func (c *Controller) AddCertificateToManage(ctx context.Context, resource *certv1.Certificate, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		mutable:   c.GlobalMutateFn(ctx),
		checkable: statuscheck.CertificateCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(res, dependencies)
}

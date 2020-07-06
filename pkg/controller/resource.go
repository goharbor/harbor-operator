package controller

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/resources/mutation"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
)

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

func (c *Controller) AddServiceToManage(ctx context.Context, resource *corev1.Service, dependencies ...graph.Resource) (graph.Resource, error) {
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

func (c *Controller) AddBasicResource(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert resource to unstuctured")
	}

	gvks, _, err := c.Scheme.ObjectKinds(resource)
	if err != nil {
		return nil, errors.Wrap(err, "cannot object kind")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)
	u.SetGroupVersionKind(gvks[0])

	return c.AddUnsctructuredToManage(ctx, u, dependencies...)
}

func (c *Controller) AddExternalResource(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.EnsureReady)
}

func (c *Controller) AddExternalTypedSecret(ctx context.Context, secret *corev1.Secret, secretType corev1.SecretType, dependencies ...graph.Resource) (graph.Resource, error) {
	if secret == nil {
		return nil, nil
	}

	resource := secret.DeepCopy()

	resource.Type = secretType

	res := &Resource{
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.EnsureReady)
}

func (c *Controller) AddCertificateToManage(ctx context.Context, resource *certv1.Certificate, dependencies ...graph.Resource) (graph.Resource, error) {
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

func (c *Controller) AddIngressToManage(ctx context.Context, resource *netv1.Ingress, dependencies ...graph.Resource) (graph.Resource, error) {
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

func (c *Controller) AddSecretToManage(ctx context.Context, resource *corev1.Secret, dependencies ...graph.Resource) (graph.Resource, error) {
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

func (c *Controller) AddConfigMapToManage(ctx context.Context, resource *corev1.ConfigMap, dependencies ...graph.Resource) (graph.Resource, error) {
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

func (c *Controller) AddDeploymentToManage(ctx context.Context, resource *appsv1.Deployment, dependencies ...graph.Resource) (graph.Resource, error) {
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

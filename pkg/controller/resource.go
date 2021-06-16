package controller

import (
	"context"

	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource struct {
	mutable   resources.Mutable
	checkable resources.Checkable
	resource  resources.Resource
}

func (res *Resource) GetResource() resources.Resource {
	return res.resource
}

func (c *Controller) Changed(ctx context.Context, depManager *checksum.Dependencies, resource resources.Resource) (bool, error) {
	objectKey := client.ObjectKeyFromObject(resource)

	result := resource.DeepCopyObject()

	// nolint:nestif
	if result, ok := result.(resources.Resource); ok {
		err := c.Client.Get(ctx, objectKey, result)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return false, errors.Wrap(err, "cannot get resource")
			}

			return true, nil
		}

		if isImmutableResource(result) {
			return false, nil
		}

		checksum.CopyVersion(result.(metav1.Object), resource)

		resultAnnotations := result.GetAnnotations()

		for key, value := range resource.GetAnnotations() {
			if resultValue, ok := resultAnnotations[key]; checksum.IsStaticAnnotation(key) && (!ok || resultValue != value) {
				return true, nil
			}
		}

		checksum.CopyMarkers(result.(metav1.Object), resource)

		return depManager.ChangedFor(ctx, resource), nil
	}

	return false, nil
}

func (c *Controller) ProcessFunc(ctx context.Context, resource runtime.Object, dependencies ...graph.Resource) func(context.Context, graph.Resource) error { // nolint:funlen
	depManager := checksum.New(c.Scheme)

	depManager.Add(ctx, owner.Get(ctx), true)

	gvks, _, err := c.Scheme.ObjectKinds(resource)
	if err == nil {
		resource.GetObjectKind().SetGroupVersionKind(gvks[0])
	}

	for _, dep := range dependencies {
		if dep, ok := dep.(*Resource); ok {
			gvks, _, err := c.Scheme.ObjectKinds(dep.resource)
			if err == nil {
				dep.resource.GetObjectKind().SetGroupVersionKind(gvks[0])
			}

			depManager.Add(ctx, dep.resource, false)
		}
	}

	return func(ctx context.Context, r graph.Resource) error {
		res, ok := r.(*Resource)
		if !ok {
			return nil
		}

		span, ctx := opentracing.StartSpanFromContext(ctx, "process")
		defer span.Finish()

		namespace, name := res.resource.GetNamespace(), res.resource.GetName()

		gvk := c.AddGVKToSpan(ctx, span, res.resource)
		l := logger.Get(ctx).WithValues(
			"resource.apiVersion", gvk.GroupVersion(),
			"resource.kind", gvk.Kind,
			"resource.name", name,
			"resource.namespace", namespace,
		)

		logger.Set(&ctx, l)
		span.
			SetTag("resource.name", name).
			SetTag("resource.namespace", namespace)

		changed, err := c.Changed(ctx, depManager, res.resource)
		if err != nil {
			return errors.Wrap(err, "changes detection")
		}

		if !changed {
			l.V(0).Info("dependencies unchanged")

			err = c.EnsureReady(ctx, res)

			return errors.Wrap(err, "check")
		}

		res.mutable.AppendMutation(func(ctx context.Context, resource runtime.Object) error {
			if res, ok := resource.(metav1.Object); ok {
				depManager.AddAnnotations(res)
			}

			return nil
		})

		err = c.applyAndCheck(ctx, r)

		return errors.Wrapf(err, "apply %s (%s/%s)", gvk, namespace, name)
	}
}

func (c *Controller) AddUnsctructuredToManage(ctx context.Context, resource *unstructured.Unstructured, dependencies ...graph.Resource) (graph.Resource, error) { // nolint:interfacer
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.UnstructuredCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddServiceToManage(ctx context.Context, resource *corev1.Service, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddBasicResource(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (*Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
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

func (c *Controller) AddExternalConfigMap(ctx context.Context, configMap *corev1.ConfigMap, dependencies ...graph.Resource) (graph.Resource, error) {
	if configMap == nil {
		return nil, nil
	}

	res := &Resource{
		checkable: statuscheck.True,
		resource:  configMap,
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

	check := statuscheck.True

	if secretType == corev1.SecretTypeTLS {
		check = statuscheck.TLSSecretCheck
	}

	res := &Resource{
		checkable: check,
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

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.CertificateCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddIssuerToManage(ctx context.Context, resource *certv1.Issuer, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.IssuerCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddIngressToManage(ctx context.Context, resource *netv1.Ingress, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddNetworkPolicyToManage(ctx context.Context, resource *netv1.NetworkPolicy, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddSecretToManage(ctx context.Context, resource *corev1.Secret, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.SecretMutateFn(ctx, resource.Immutable)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddConfigMapToManage(ctx context.Context, resource *corev1.ConfigMap, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.True,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddDeploymentToManage(ctx context.Context, resource *appsv1.Deployment, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	mutate, err := c.DeploymentMutateFn(ctx, dependencies...)
	if err != nil {
		return nil, err
	}

	res := &Resource{
		mutable:   mutate,
		checkable: statuscheck.BasicCheck,
		resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

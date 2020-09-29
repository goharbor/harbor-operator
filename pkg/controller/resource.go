package controller

import (
	"context"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/goharbor/harbor-operator/pkg/resources/mutation"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Resource struct {
	Mutable   resources.Mutable
	Checkable resources.Checkable
	Resource  resources.Resource
}

func (c *Controller) SetGVK(ctx context.Context, resource resources.Resource) error {
	gvks, _, err := c.Scheme.ObjectKinds(resource)
	if err != nil {
		return errors.Wrap(err, "groupVersionKind")
	}

	resource.SetGroupVersionKind(gvks[0])

	return nil
}

func (c *Controller) ProcessFunc(ctx context.Context, resource metav1.Object, dependencies ...graph.Resource) func(context.Context, graph.Resource) error { // nolint:funlen
	depManager := checksum.New(c.Scheme)

	depManager.Add(ctx, owner.Get(ctx), false)

	for _, dep := range dependencies {
		if dep, ok := dep.(*Resource); ok {
			depManager.Add(ctx, dep.Resource, true)
		}
	}

	return func(ctx context.Context, r graph.Resource) error {
		res, ok := r.(*Resource)
		if !ok {
			return nil
		}

		span, ctx := opentracing.StartSpanFromContext(ctx, "process")
		defer span.Finish()

		namespace, name := res.Resource.GetNamespace(), res.Resource.GetName()

		gvk := c.AddGVKToSpan(ctx, span, res.Resource)
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

		objectKey, err := client.ObjectKeyFromObject(res.Resource)
		if err != nil {
			return serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get object key")
		}

		result := res.Resource.DeepCopyObject()

		err = c.Client.Get(ctx, objectKey, result)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return errors.Wrap(err, "cannot get resource")
			}
		} else {
			checksum.CopyMarkers(result.(metav1.Object), res.Resource)
		}

		if !depManager.ChangedFor(ctx, res.Resource) {
			changed := false

			for key := range res.Resource.GetAnnotations() {
				if checksum.IsStaticAnnotation(key) {
					changed = true

					break
				}
			}

			if !changed {
				l.V(0).Info("dependencies unchanged")

				return nil
			}
		}

		res.Mutable.AppendMutation(func(ctx context.Context, resource, result runtime.Object) controllerutil.MutateFn {
			return func() error {
				if res, ok := result.(metav1.Object); ok {
					depManager.AddAnnotations(res)
					depManager.AddAnnotations(r.(*Resource).Resource)
				}

				return nil
			}
		})

		err = c.applyAndCheck(ctx, r)

		return errors.Wrapf(err, "apply %s (%s/%s)", gvk, namespace, name)
	}
}

func (c *Controller) AddUnsctructuredToManage(ctx context.Context, resource *unstructured.Unstructured, dependencies ...graph.Resource) (graph.Resource, error) { // nolint:interfacer
	if resource == nil {
		return nil, nil
	}

	res := &Resource{
		Mutable:   mutation.NewUnstructured(c.GlobalMutateFn(ctx)),
		Checkable: statuscheck.UnstructuredCheck,
		Resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddServiceToManage(ctx context.Context, service *corev1.Service, dependencies ...graph.Resource) (graph.Resource, error) {
	if service == nil {
		return nil, nil
	}

	resource := service.DeepCopyObject().(*corev1.Service)

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable: mutation.NewService(c.GlobalMutateFn(ctx)),
		Checkable: func(ctx context.Context, object runtime.Object) (bool, error) {
			service := object.(*corev1.Service)

			ok, err := statuscheck.BasicCheck(ctx, service)
			if err != nil || !ok {
				return ok, err
			}

			var endpoint corev1.Endpoints

			c.Client.Get(ctx, types.NamespacedName{
				Namespace: service.GetNamespace(),
				Name:      service.GetName(),
			}, &endpoint)

			ports := make([]string, len(service.Spec.Ports))

			for i, port := range service.Spec.Ports {
				ports[i] = port.Name
			}

			return statuscheck.EndpointCheck(ctx, &endpoint, ports...)
		},
		Resource: resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddBasicResource(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert resource to unstuctured")
	}

	err = c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	return c.AddUnsctructuredToManage(ctx, u, dependencies...)
}

func (c *Controller) AddExternalResource(ctx context.Context, resource resources.Resource, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Checkable: statuscheck.BasicCheck,
		Resource:  resource,
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

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	resource.Type = secretType

	res := &Resource{
		Checkable: statuscheck.BasicCheck,
		Resource:  resource,
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

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable:   mutation.NewCertificate(c.GlobalMutateFn(ctx)),
		Checkable: statuscheck.CertificateCheck,
		Resource:  resource,
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

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable:   mutation.NewIssuer(c.GlobalMutateFn(ctx)),
		Checkable: statuscheck.IssuerCheck,
		Resource:  resource,
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

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable:   mutation.NewIngress(c.GlobalMutateFn(ctx)),
		Checkable: statuscheck.IngressCheck,
		Resource:  resource,
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

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable:   mutation.NewSecret(c.GlobalMutateFn(ctx), true, false),
		Checkable: statuscheck.True,
		Resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddImmutableSecretToManage(ctx context.Context, resource *corev1.Secret, dependencies ...graph.Resource) (graph.Resource, error) {
	if resource == nil {
		return nil, nil
	}

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable:   mutation.NewSecret(c.GlobalMutateFn(ctx), false, false),
		Checkable: statuscheck.True,
		Resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddConfigMapToManage(ctx context.Context, configmap *corev1.ConfigMap, dependencies ...graph.Resource) (graph.Resource, error) {
	if configmap == nil {
		return nil, nil
	}

	resource := configmap.DeepCopyObject().(*corev1.ConfigMap)

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable:   mutation.NewConfigMap(c.GlobalMutateFn(ctx)),
		Checkable: statuscheck.True,
		Resource:  resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

func (c *Controller) AddDeploymentToManage(ctx context.Context, deployment *appsv1.Deployment, dependencies ...graph.Resource) (graph.Resource, error) {
	if deployment == nil {
		return nil, nil
	}

	resource := deployment.DeepCopyObject().(*appsv1.Deployment)

	err := c.SetGVK(ctx, resource)
	if err != nil {
		return nil, errors.Wrap(err, "gvk")
	}

	res := &Resource{
		Mutable: mutation.NewDeployment(c.DeploymentMutateFn(ctx, dependencies...)),
		Checkable: func(ctx context.Context, object runtime.Object) (bool, error) {
			return statuscheck.BasicCheck(ctx, object)
		},
		Resource: resource,
	}

	g := sgraph.Get(ctx)
	if g == nil {
		return nil, errors.Errorf("no graph in current context")
	}

	return res, g.AddResource(ctx, res, dependencies, c.ProcessFunc(ctx, resource, dependencies...))
}

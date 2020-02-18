package components

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	harbor_chartmuseum "github.com/ovh/harbor-operator/controllers/harbor/components/chartmuseum"
	harbor_clair "github.com/ovh/harbor-operator/controllers/harbor/components/clair"
	harbor_core "github.com/ovh/harbor-operator/controllers/harbor/components/harbor-core"
	harbor_jobservice "github.com/ovh/harbor-operator/controllers/harbor/components/jobservice"
	harbor_notary "github.com/ovh/harbor-operator/controllers/harbor/components/notary"
	harbor_portal "github.com/ovh/harbor-operator/controllers/harbor/components/portal"
	harbor_registry "github.com/ovh/harbor-operator/controllers/harbor/components/registry"
	"github.com/ovh/harbor-operator/pkg/factories/logger"
)

type Resource interface {
	metav1.Object
	runtime.Object
	schema.ObjectKind
}

type Components struct {
	Core        *ComponentRunner
	JobService  *ComponentRunner
	Registry    *ComponentRunner
	Portal      *ComponentRunner
	ChartMuseum *ComponentRunner
	Clair       *ComponentRunner
	Notary      *ComponentRunner
}

type Component interface {
	GetConfigMaps(context.Context) []*corev1.ConfigMap
	GetSecrets(context.Context) []*corev1.Secret
	GetServices(context.Context) []*corev1.Service
	GetCertificates(context.Context) []*certv1.Certificate
	GetIngresses(context.Context) []*netv1.Ingress
	GetDeployments(context.Context) []*appsv1.Deployment
}

func GetComponents(ctx context.Context, harbor *containerregistryv1alpha1.Harbor) (*Components, error) {
	harborResource := &Components{}

	var g errgroup.Group

	g.Go(harborResource.Core.getInitFunc(ctx, harbor, CorePriority, containerregistryv1alpha1.CoreName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
		return harbor_core.New(ctx, harbor, option)
	}))
	g.Go(harborResource.Registry.getInitFunc(ctx, harbor, RegistryPriority, containerregistryv1alpha1.RegistryName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
		return harbor_registry.New(ctx, harbor, option)
	}))
	g.Go(harborResource.Portal.getInitFunc(ctx, harbor, PortalPriority, containerregistryv1alpha1.PortalName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
		return harbor_portal.New(ctx, harbor, option)
	}))
	g.Go(harborResource.JobService.getInitFunc(ctx, harbor, JobServicePriority, containerregistryv1alpha1.JobServiceName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
		return harbor_jobservice.New(ctx, harbor, option)
	}))

	if harbor.Spec.Components.ChartMuseum != nil {
		g.Go(harborResource.ChartMuseum.getInitFunc(ctx, harbor, ChartMuseumPriority, containerregistryv1alpha1.ChartMuseumName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
			return harbor_chartmuseum.New(ctx, harbor, option)
		}))
	}

	if harbor.Spec.Components.Clair != nil {
		g.Go(harborResource.Clair.getInitFunc(ctx, harbor, ClairPriority, containerregistryv1alpha1.ClairName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
			return harbor_clair.New(ctx, harbor, option)
		}))
	}

	if harbor.Spec.Components.Notary != nil {
		g.Go(harborResource.Notary.getInitFunc(ctx, harbor, NotaryPriority, containerregistryv1alpha1.NotaryName, func(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, option *Option) (Component, error) {
			return harbor_notary.New(ctx, harbor, option)
		}))
	}

	err := g.Wait()

	return harborResource, errors.Wrap(err, "cannot get resources")
}

type ComponentFactory func(context.Context, *containerregistryv1alpha1.Harbor, OptionGetter) (Component, error)

func (c *ComponentRunner) getOption(harbor *containerregistryv1alpha1.Harbor, componentPriority int32) *Option {
	option := &Option{}

	if harbor.Spec.Priority != nil {
		priority := *harbor.Spec.Priority - PriorityBase + componentPriority
		option.SetPriority(&priority)
	}

	return option
}

func (c *ComponentRunner) getInitFunc(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, componentPriority int32, name string, factory func(context.Context, *containerregistryv1alpha1.Harbor, *Option) (Component, error)) func() error {
	return func() error {
		options := c.getOption(harbor, componentPriority)

		ctx := withComponent(ctx, name)

		span, ctx := opentracing.StartSpanFromContext(ctx, "init", opentracing.Tags{
			"component": name,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Component", name))

		component, err := factory(ctx, harbor, options)
		if err != nil {
			return errors.Wrap(err, name)
		}

		c.Component = component

		return nil
	}
}

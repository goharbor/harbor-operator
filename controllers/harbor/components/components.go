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
	harbor_portal "github.com/ovh/harbor-operator/controllers/harbor/components/portal"
	harbor_registry "github.com/ovh/harbor-operator/controllers/harbor/components/registry"
	"github.com/ovh/harbor-operator/pkg/factories/logger"
)

const PriorityBase = 100

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
}

type ComponentRunner struct {
	Component
}

type Component interface {
	GetConfigMaps(context.Context) []*corev1.ConfigMap
	GetSecrets(context.Context) []*corev1.Secret
	GetServices(context.Context) []*corev1.Service
	GetCertificates(context.Context) []*certv1.Certificate
	GetIngresses(context.Context) []*netv1.Ingress
	GetDeployments(context.Context) []*appsv1.Deployment
}

func GetComponents(ctx context.Context, harbor *containerregistryv1alpha1.Harbor) (*Components, error) { // nolint:funlen
	harborResource := &Components{}

	var g errgroup.Group

	g.Go(func() error {
		var corePriority *int32
		if harbor.Spec.Priority != nil {
			priority := *harbor.Spec.Priority - PriorityBase + CorePriority
			corePriority = &priority
		}

		core, err := harbor_core.New(ctx, harbor, harbor_core.Option{Priority: corePriority})
		if err != nil {
			return errors.Wrap(err, containerregistryv1alpha1.CoreName)
		}
		harborResource.Core = &ComponentRunner{core}
		return nil
	})

	g.Go(func() error {
		var registryPriority *int32
		if harbor.Spec.Priority != nil {
			priority := *harbor.Spec.Priority - PriorityBase + RegistryPriority
			registryPriority = &priority
		}

		reg, err := harbor_registry.New(ctx, harbor, harbor_registry.Option{Priority: registryPriority})
		if err != nil {
			return errors.Wrap(err, containerregistryv1alpha1.RegistryName)
		}
		harborResource.Registry = &ComponentRunner{reg}
		return nil
	})

	g.Go(func() error {
		var portalPriority *int32
		if harbor.Spec.Priority != nil {
			priority := *harbor.Spec.Priority - PriorityBase + PortalPriority
			portalPriority = &priority
		}

		portal, err := harbor_portal.New(ctx, harbor, harbor_portal.Option{Priority: portalPriority})
		if err != nil {
			return errors.Wrap(err, containerregistryv1alpha1.PortalName)
		}
		harborResource.Portal = &ComponentRunner{portal}
		return nil
	})

	g.Go(func() error {
		var jobServicePriority *int32
		if harbor.Spec.Priority != nil {
			priority := *harbor.Spec.Priority - PriorityBase + JobServicePriority
			jobServicePriority = &priority
		}

		jobService, err := harbor_jobservice.New(ctx, harbor, harbor_jobservice.Option{Priority: jobServicePriority})
		if err != nil {
			return errors.Wrap(err, containerregistryv1alpha1.JobServiceName)
		}
		harborResource.JobService = &ComponentRunner{jobService}
		return nil
	})

	if harbor.Spec.Components.ChartMuseum != nil {
		g.Go(func() error {
			var chartMuseumPriority *int32
			if harbor.Spec.Priority != nil {
				priority := *harbor.Spec.Priority - PriorityBase + ChartMuseumPriority
				chartMuseumPriority = &priority
			}

			chartMuseum, err := harbor_chartmuseum.New(ctx, harbor, harbor_chartmuseum.Option{Priority: chartMuseumPriority})
			if err != nil {
				return errors.Wrap(err, containerregistryv1alpha1.ChartMuseumName)
			}
			harborResource.ChartMuseum = &ComponentRunner{chartMuseum}
			return nil
		})
	}

	if harbor.Spec.Components.Clair != nil {
		g.Go(func() error {
			var clairPriority *int32
			if harbor.Spec.Priority != nil {
				priority := *harbor.Spec.Priority - PriorityBase + ClairPriority
				clairPriority = &priority
			}

			clair, err := harbor_clair.New(ctx, harbor, harbor_clair.Option{Priority: clairPriority})
			if err != nil {
				return errors.Wrap(err, containerregistryv1alpha1.ClairName)
			}
			harborResource.Clair = &ComponentRunner{clair}
			return nil
		})
	}

	err := g.Wait()

	return harborResource, errors.Wrap(err, "cannot get resources")
}

type Run func(context.Context, *containerregistryv1alpha1.Harbor, *ComponentRunner) error

// nolint:funlen
func (r *Components) ParallelRun(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, run Run) error {
	var g errgroup.Group

	g.Go(func() error {
		ctx := withComponent(ctx, containerregistryv1alpha1.CoreName)
		span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
			"component": containerregistryv1alpha1.CoreName,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Component", containerregistryv1alpha1.CoreName))

		err := run(ctx, harbor, r.Core)
		return errors.Wrap(err, containerregistryv1alpha1.CoreName)
	})

	g.Go(func() error {
		ctx := withComponent(ctx, containerregistryv1alpha1.RegistryName)
		span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
			"component": containerregistryv1alpha1.RegistryName,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Component", containerregistryv1alpha1.RegistryName))

		err := run(ctx, harbor, r.Registry)
		return errors.Wrap(err, containerregistryv1alpha1.RegistryName)
	})

	g.Go(func() error {
		ctx := withComponent(ctx, containerregistryv1alpha1.JobServiceName)
		span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
			"component": containerregistryv1alpha1.JobServiceName,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Component", containerregistryv1alpha1.JobServiceName))

		err := run(ctx, harbor, r.JobService)
		return errors.Wrap(err, containerregistryv1alpha1.JobServiceName)
	})

	g.Go(func() error {
		ctx := withComponent(ctx, containerregistryv1alpha1.PortalName)
		span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
			"component": containerregistryv1alpha1.PortalName,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Component", containerregistryv1alpha1.PortalName))

		err := run(ctx, harbor, r.Portal)
		return errors.Wrap(err, containerregistryv1alpha1.PortalName)
	})

	if r.ChartMuseum != nil {
		g.Go(func() error {
			ctx := withComponent(ctx, containerregistryv1alpha1.ChartMuseumName)
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"component": containerregistryv1alpha1.ChartMuseumName,
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Component", containerregistryv1alpha1.ChartMuseumName))

			err := run(ctx, harbor, r.ChartMuseum)
			return errors.Wrap(err, containerregistryv1alpha1.ChartMuseumName)
		})
	}

	if r.Clair != nil {
		g.Go(func() error {
			ctx := withComponent(ctx, containerregistryv1alpha1.ClairName)
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"component": containerregistryv1alpha1.ClairName,
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Component", containerregistryv1alpha1.ClairName))

			err := run(ctx, harbor, r.Clair)
			return errors.Wrap(err, containerregistryv1alpha1.ClairName)
		})
	}

	return g.Wait()
}

type ComponentRun func(context.Context, *containerregistryv1alpha1.Harbor, []Resource) error

// ParallelRun run a function over all resources of a component.
// This is a wrapper which use errgroup.
// The main goal of this method is to centralize action over Resource
// and not forget any resources anywhere else in the code.
func (c *ComponentRunner) ParallelRun(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, servicesRun, configMapsRun, ingressesRun, secretRun, certificatesRun, deploymentsRun ComponentRun, waitBeforeDeployments bool) error { // nolint:funlen
	if c == nil {
		return nil
	}

	var g errgroup.Group

	if servicesRun != nil {
		g.Go(func() error {
			services := c.GetServices(ctx)
			resources := make([]Resource, len(services))
			for i, d := range services {
				resources[i] = d
			}

			ctx := withResource(ctx, "services")
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"Resource.Kind": "services",
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", "services"))

			err := servicesRun(ctx, harbor, resources)
			return errors.Wrap(err, "services")
		})
	}

	if configMapsRun != nil {
		g.Go(func() error {
			configmaps := c.GetConfigMaps(ctx)
			resources := make([]Resource, len(configmaps))
			for i, d := range configmaps {
				resources[i] = d
			}

			ctx := withResource(ctx, "configmaps")
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"Resource.Kind": "configmaps",
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", "configmaps"))

			err := configMapsRun(ctx, harbor, resources)
			return errors.Wrap(err, "configmaps")
		})
	}

	if ingressesRun != nil {
		g.Go(func() error {
			ingresses := c.GetIngresses(ctx)
			resources := make([]Resource, len(ingresses))
			for i, d := range ingresses {
				resources[i] = d
			}

			ctx := withResource(ctx, "ingresses")
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"Resource.Kind": "ingresses",
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", "ingresses"))

			err := ingressesRun(ctx, harbor, resources)
			return errors.Wrap(err, "ingresses")
		})
	}

	if secretRun != nil {
		g.Go(func() error {
			secrets := c.GetSecrets(ctx)
			resources := make([]Resource, len(secrets))
			for i, d := range secrets {
				resources[i] = d
			}

			ctx := withResource(ctx, "secrets")
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"Resource.Kind": "secrets",
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", "secrets"))

			err := secretRun(ctx, harbor, resources)
			return errors.Wrap(err, "secrets")
		})
	}

	if certificatesRun != nil {
		g.Go(func() error {
			certificates := c.GetCertificates(ctx)
			resources := make([]Resource, len(certificates))
			for i, d := range certificates {
				resources[i] = d
			}

			ctx := withResource(ctx, "certificates")
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"Resource.Kind": "certificates",
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", "certificates"))

			err := certificatesRun(ctx, harbor, resources)
			return errors.Wrap(err, "certificates")
		})
	}

	if waitBeforeDeployments {
		err := g.Wait()
		if err != nil {
			return err
		}
	}

	if deploymentsRun != nil {
		g.Go(func() error {
			deployments := c.GetDeployments(ctx)
			resources := make([]Resource, len(deployments))
			for i, d := range deployments {
				resources[i] = d
			}

			ctx := withResource(ctx, "deployments")
			span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
				"Resource.Kind": "deployments",
			})
			defer span.Finish()

			logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", "deployments"))

			err := deploymentsRun(ctx, harbor, resources)
			return errors.Wrap(err, "deployments")
		})
	}

	return g.Wait()
}

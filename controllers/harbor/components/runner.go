package components

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/logger"
)

type ComponentRunner struct {
	Component
}

type Run func(context.Context, *containerregistryv1alpha1.Harbor, *ComponentRunner) error

func (r *Components) ParallelRun(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, run Run) error {
	var g errgroup.Group

	g.Go(run.getRunFunc(ctx, harbor, r.Core, containerregistryv1alpha1.CoreName))
	g.Go(run.getRunFunc(ctx, harbor, r.Registry, containerregistryv1alpha1.RegistryName))
	g.Go(run.getRunFunc(ctx, harbor, r.JobService, containerregistryv1alpha1.JobServiceName))
	g.Go(run.getRunFunc(ctx, harbor, r.Portal, containerregistryv1alpha1.PortalName))
	g.Go(run.getRunFunc(ctx, harbor, r.ChartMuseum, containerregistryv1alpha1.ChartMuseumName))
	g.Go(run.getRunFunc(ctx, harbor, r.Clair, containerregistryv1alpha1.ClairName))
	g.Go(run.getRunFunc(ctx, harbor, r.Notary, containerregistryv1alpha1.NotaryName))

	return g.Wait()
}

func (r Run) getRunFunc(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, runner *ComponentRunner, name string) func() error {
	return func() error {
		if runner == nil {
			return nil
		}

		ctx := withComponent(ctx, name)

		span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
			"component": name,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Component", name))

		return errors.Wrap(r(ctx, harbor, runner), name)
	}
}

type ComponentRun func(context.Context, *containerregistryv1alpha1.Harbor, []Resource) error

// ParallelRun run a function over all resources of a component.
// This is a wrapper which use errgroup.
// The main goal of this method is to centralize action over Resource
// and not forget any resources anywhere else in the code.
func (c *ComponentRunner) ParallelRun(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, servicesRun, configMapsRun, ingressesRun, secretsRun, certificatesRun, deploymentsRun ComponentRun, waitBeforeDeployments bool) error {
	if c == nil {
		return nil
	}

	var g errgroup.Group

	g.Go(servicesRun.getRunFunc(ctx, harbor, c.GetServices(ctx), "services"))
	g.Go(configMapsRun.getRunFunc(ctx, harbor, c.GetConfigMaps(ctx), "configmaps"))
	g.Go(ingressesRun.getRunFunc(ctx, harbor, c.GetIngresses(ctx), "ingresses"))
	g.Go(secretsRun.getRunFunc(ctx, harbor, c.GetSecrets(ctx), "secrets"))
	g.Go(certificatesRun.getRunFunc(ctx, harbor, c.GetCertificates(ctx), "certificates"))

	if waitBeforeDeployments {
		err := g.Wait()
		if err != nil {
			return err
		}
	}

	g.Go(deploymentsRun.getRunFunc(ctx, harbor, c.GetDeployments(ctx), "deployments"))

	return g.Wait()
}

func (c *ComponentRun) getRunFunc(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, resources []Resource, kind string) func() error {
	return func() error {
		if c == nil {
			return nil
		}

		ctx := withResource(ctx, kind)

		span, ctx := opentracing.StartSpanFromContext(ctx, "run", opentracing.Tags{
			"Resource.Kind": kind,
		})
		defer span.Finish()

		logger.Set(&ctx, logger.Get(ctx).WithValues("Resource.Kind", kind))

		return errors.Wrap((*c)(ctx, harbor, resources), kind)
	}
}

func (c *ComponentRunner) GetServices(ctx context.Context) []Resource {
	services := c.Component.GetServices(ctx)

	resources := make([]Resource, len(services))
	for i, r := range services {
		resources[i] = r
	}

	return resources
}

func (c *ComponentRunner) GetConfigMaps(ctx context.Context) []Resource {
	configmaps := c.Component.GetConfigMaps(ctx)

	resources := make([]Resource, len(configmaps))
	for i, r := range configmaps {
		resources[i] = r
	}

	return resources
}

func (c *ComponentRunner) GetIngresses(ctx context.Context) []Resource {
	ingresses := c.Component.GetIngresses(ctx)

	resources := make([]Resource, len(ingresses))
	for i, r := range ingresses {
		resources[i] = r
	}

	return resources
}

func (c *ComponentRunner) GetSecrets(ctx context.Context) []Resource {
	secrets := c.Component.GetSecrets(ctx)

	resources := make([]Resource, len(secrets))
	for i, r := range secrets {
		resources[i] = r
	}

	return resources
}

func (c *ComponentRunner) GetCertificates(ctx context.Context) []Resource {
	certificates := c.Component.GetCertificates(ctx)

	resources := make([]Resource, len(certificates))
	for i, r := range certificates {
		resources[i] = r
	}

	return resources
}

func (c *ComponentRunner) GetDeployments(ctx context.Context) []Resource {
	deployments := c.Component.GetDeployments(ctx)

	resources := make([]Resource, len(deployments))
	for i, r := range deployments {
		resources[i] = r
	}

	return resources
}

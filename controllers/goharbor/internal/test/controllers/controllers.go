package controllers

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/jobservice"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notarysigner"
	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registryctl"
	"github.com/goharbor/harbor-operator/controllers/goharbor/trivy"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/onsi/gomega"
	"github.com/ovh/configstore"
)

func NewCore(ctx context.Context, className string) *core.Reconciler {
	name := controllers.Core.String()

	return New(ctx, name, className, core.New).(*core.Reconciler)
}

func NewChartMuseum(ctx context.Context, className string) *chartmuseum.Reconciler {
	name := controllers.ChartMuseum.String()

	return New(ctx, name, className, chartmuseum.New).(*chartmuseum.Reconciler)
}

func NewTrivy(ctx context.Context, className string) *trivy.Reconciler {
	name := controllers.Trivy.String()

	return New(ctx, name, className, trivy.New).(*trivy.Reconciler)
}

func NewNotaryServer(ctx context.Context, className string) *notaryserver.Reconciler {
	name := controllers.NotaryServer.String()

	return New(ctx, name, className, notaryserver.New).(*notaryserver.Reconciler)
}

func NewNotarySigner(ctx context.Context, className string) *notarysigner.Reconciler {
	name := controllers.NotarySigner.String()

	return New(ctx, name, className, notarysigner.New).(*notarysigner.Reconciler)
}

func NewJobService(ctx context.Context, className string) *jobservice.Reconciler {
	name := controllers.JobService.String()

	return New(ctx, name, className, jobservice.New).(*jobservice.Reconciler)
}

func NewPortal(ctx context.Context, className string) *portal.Reconciler {
	name := controllers.Portal.String()

	return New(ctx, name, className, portal.New).(*portal.Reconciler)
}

func NewRegistry(ctx context.Context, className string) *registry.Reconciler {
	name := controllers.Registry.String()

	return New(ctx, name, className, registry.New).(*registry.Reconciler)
}

func NewRegistryCtl(ctx context.Context, className string) *registryctl.Reconciler {
	name := controllers.RegistryController.String()

	return New(ctx, name, className, registryctl.New).(*registryctl.Reconciler)
}

type CtrlBuilder func(context.Context, *configstore.Store) (controller.Reconciler, error)

func New(ctx context.Context, name, className string, builder CtrlBuilder) controller.Reconciler {
	configStore := config.NewConfigWithDefaults()
	provider := configStore.InMemory("test")
	provider.Add(configstore.NewItem(config.HarborClassKey, className, 100))
	configStore.Env(name)

	r, err := builder(ctx, configStore)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	mgr := test.GetManager(ctx)

	gomega.Expect(r.SetupWithManager(ctx, mgr)).
		To(gomega.Succeed())

	return r
}

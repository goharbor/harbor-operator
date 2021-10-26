package controllers

import (
	"context"
	"path"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/jobservice"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notarysigner"
	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/controllers/goharbor/trivy"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/setup"
	"github.com/onsi/gomega"
	"github.com/ovh/configstore"
)

const configDirectory = "../../../config/config"

func NewCore(ctx context.Context, className string) *core.Reconciler {
	return New(ctx, controllers.Core, className, core.New).(*core.Reconciler)
}

func NewChartMuseum(ctx context.Context, className string) *chartmuseum.Reconciler {
	return New(ctx, controllers.ChartMuseum, className, chartmuseum.New).(*chartmuseum.Reconciler)
}

func NewTrivy(ctx context.Context, className string) *trivy.Reconciler {
	return New(ctx, controllers.Trivy, className, trivy.New).(*trivy.Reconciler)
}

func NewNotaryServer(ctx context.Context, className string) *notaryserver.Reconciler {
	return New(ctx, controllers.NotaryServer, className, notaryserver.New).(*notaryserver.Reconciler)
}

func NewNotarySigner(ctx context.Context, className string) *notarysigner.Reconciler {
	return New(ctx, controllers.NotarySigner, className, notarysigner.New).(*notarysigner.Reconciler)
}

func NewJobService(ctx context.Context, className string) *jobservice.Reconciler {
	return New(ctx, controllers.JobService, className, jobservice.New).(*jobservice.Reconciler)
}

func NewPortal(ctx context.Context, className string) *portal.Reconciler {
	return New(ctx, controllers.Portal, className, portal.New).(*portal.Reconciler)
}

func NewRegistry(ctx context.Context, className string) *registry.Reconciler {
	return New(ctx, controllers.Registry, className, registry.New).(*registry.Reconciler)
}

type CtrlBuilder func(context.Context, *configstore.Store) (controller.Reconciler, error)

func New(ctx context.Context, name controllers.Controller, className string, builder CtrlBuilder) controller.Reconciler {
	ctrl := setup.NewController(name, builder)

	configstore.InMemory(test.NewName("test-")).
		Add(configstore.NewItem(config.HarborClassKey, className, 100)).
		Add(configstore.NewItem(config.TemplateDirectoryKey, path.Join(configDirectory, "assets"), 100)).
		Add(configstore.NewItem(config.CtrlConfigDirectoryKey, path.Join(configDirectory, "controllers"), 100))

	mgr := test.GetManager(ctx)

	reconciler, err := ctrl.WithManager(ctx, mgr)
	gomega.Expect(err).
		ToNot(gomega.HaveOccurred())

	return reconciler
}

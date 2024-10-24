package controllers

import (
	"context"
	"path"

	"github.com/onsi/gomega"
	"github.com/ovh/configstore"
	"github.com/plotly/harbor-operator/controllers"
	"github.com/plotly/harbor-operator/controllers/goharbor/core"
	"github.com/plotly/harbor-operator/controllers/goharbor/internal/test"
	"github.com/plotly/harbor-operator/controllers/goharbor/jobservice"
	"github.com/plotly/harbor-operator/controllers/goharbor/portal"
	"github.com/plotly/harbor-operator/controllers/goharbor/registry"
	"github.com/plotly/harbor-operator/controllers/goharbor/trivy"
	"github.com/plotly/harbor-operator/pkg/config"
	"github.com/plotly/harbor-operator/pkg/controller"
	"github.com/plotly/harbor-operator/pkg/setup"
)

const configDirectory = "../../../config/config"

func NewCore(ctx context.Context, className string) *core.Reconciler {
	return New(ctx, controllers.Core, className, core.New).(*core.Reconciler)
}

func NewTrivy(ctx context.Context, className string) *trivy.Reconciler {
	return New(ctx, controllers.Trivy, className, trivy.New).(*trivy.Reconciler)
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

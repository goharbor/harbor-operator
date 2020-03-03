package setup

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/goharbor/harbor-operator/pkg/controllers"
	"github.com/goharbor/harbor-operator/pkg/controllers/chartmuseum"
	"github.com/goharbor/harbor-operator/pkg/controllers/core"
	"github.com/goharbor/harbor-operator/pkg/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/controllers/jobservice"
	"github.com/goharbor/harbor-operator/pkg/controllers/portal"
	"github.com/goharbor/harbor-operator/pkg/controllers/registry"
)

type ControllerBuilder func(context.Context, string) (controllers.Controller, error)

func SetupWithManager(ctx context.Context, mgr manager.Manager, version string) error {
	var g errgroup.Group

	g.Go(ControllerFactory(ctx, mgr, chartmuseum.New, version))
	g.Go(ControllerFactory(ctx, mgr, core.New, version))
	g.Go(ControllerFactory(ctx, mgr, harbor.New, version))
	g.Go(ControllerFactory(ctx, mgr, jobservice.New, version))
	g.Go(ControllerFactory(ctx, mgr, registry.New, version))
	g.Go(ControllerFactory(ctx, mgr, portal.New, version))

	return g.Wait()
}

func ControllerFactory(ctx context.Context, mgr manager.Manager, factory ControllerBuilder, version string) func() error {
	return func() error {
		controller, err := factory(ctx, version)
		if err != nil {
			return errors.Wrap(err, "create")
		}

		err = controller.SetupWithManager(mgr)
		return errors.Wrap(err, "setup")
	}
}

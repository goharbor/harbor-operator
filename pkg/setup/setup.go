package setup

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func WithManager(ctx context.Context, mgr manager.Manager) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return ControllersWithManager(ctx, mgr)
	})

	g.Go(func() error {
		return WebhooksWithManager(ctx, mgr)
	})

	return g.Wait()
}

func ControllersWithManager(ctx context.Context, mgr manager.Manager) error {
	var g errgroup.Group

	for name, builder := range controllersBuilder {
		name := name
		c := &controller{
			Name: name,
			New:  builder,
		}

		ok, err := c.IsEnabled(ctx)
		if err != nil {
			return errors.Wrapf(err, "cannot check if controller %s is enabled", name)
		}

		if !ok {
			logger.Get(ctx).Info("Controller disabled", "controller", name)

			continue
		}

		g.Go(func() error {
			return errors.Wrapf(c.WithManager(ctx, mgr), "controller %s", name)
		})
	}

	return g.Wait()
}

func WebhooksWithManager(ctx context.Context, mgr manager.Manager) error {
	var g errgroup.Group

	for name, object := range webhooksBuilder {
		name := name
		wh := &webHook{
			Name:    name,
			webhook: object,
		}

		ok, err := wh.IsEnabled(ctx)
		if err != nil {
			return errors.Wrapf(err, "cannot check if webhook %s is enabled", name)
		}

		if !ok {
			logger.Get(ctx).Info("Controller disabled", "controller", name)

			continue
		}

		g.Go(func() error {
			return errors.Wrapf(wh.WithManager(ctx, mgr), "webhook %s", name)
		})
	}

	return g.Wait()
}

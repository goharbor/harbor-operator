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
		return errors.Wrap(ControllersWithManager(ctx, mgr), "controllers")
	})

	g.Go(func() error {
		return errors.Wrap(WebhooksWithManager(ctx, mgr), "webhooks")
	})

	return g.Wait()
}

func ControllersWithManager(ctx context.Context, mgr manager.Manager) error {
	var g errgroup.Group

	for name, builder := range controllersBuilder {
		ctx := ctx

		logger.Set(&ctx, logger.Get(ctx).WithName(name.String()))

		c := NewController(name, builder)

		ok, err := c.IsEnabled(ctx)
		if err != nil {
			return errors.Wrap(err, "cannot check if controller is enabled")
		}

		if !ok {
			logger.Get(ctx).Info("Controller disabled")

			continue
		}

		name := name

		g.Go(func() error {
			_, err := c.WithManager(ctx, mgr)

			return errors.Wrap(err, name.String())
		})
	}

	return g.Wait()
}

func WebhooksWithManager(ctx context.Context, mgr manager.Manager) error {
	for name, object := range webhooksBuilder {
		ctx := ctx

		logger.Set(&ctx, logger.Get(ctx).WithName(name.String()))

		wh := &webHook{
			Name:    name,
			webhook: object,
		}

		ok, err := wh.IsEnabled(ctx)
		if err != nil {
			return errors.Wrap(err, "cannot check if webhook is enabled")
		}

		if !ok {
			logger.Get(ctx).Info("Webhook disabled")

			continue
		}

		// Fail earlier.
		// 'controller-runtime' does not support webhook registrations concurrently.
		// Check issue: https://github.com/kubernetes-sigs/controller-runtime/issues/1285.
		if err := wh.WithManager(ctx, mgr); err != nil {
			return errors.Wrap(err, name.String())
		}
	}

	return nil
}

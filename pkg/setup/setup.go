package setup

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

func WithManager(ctx context.Context, mgr manager.Manager) error {
	var g errgroup.Group

	for name, builder := range controllers {
		c := &controller{
			Name: name,
			New:  builder,
		}

		ok, err := c.IsEnabled(ctx)
		if err != nil {
			return errors.Wrapf(err, "controller %s", name)
		}

		if !ok {
			logger.Get(ctx).Info("Controller disabled", "controller", name)

			continue
		}

		g.Go(func() error {
			return errors.Wrapf(c.WithManager(ctx, mgr), "%s", c.Name)
		})
	}

	for name, object := range webhooks {
		wh := &webHook{
			Name:    name,
			webhook: object,
		}

		ok, err := wh.IsEnabled(ctx)
		if err != nil {
			return errors.Wrapf(err, "controller %s", name)
		}

		if !ok {
			logger.Get(ctx).Info("Controller disabled", "controller", name)
			continue
		}

		g.Go(func() error {
			return errors.Wrapf(wh.WithManager(ctx, mgr), "%s", wh.Name)
		})
	}

	return g.Wait()
}

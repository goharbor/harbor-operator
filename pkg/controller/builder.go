package controller

import (
	"context"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
)

func (c *Controller) Builder(ctx context.Context, mgr ctrl.Manager) (*builder.Builder, error) {
	className, err := c.ConfigStore.GetItemValue(config.HarborClassKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, errors.Wrap(err, "cannot get harbor class")
		}
	}

	concurrentReconcile, err := c.ConfigStore.GetItemValueInt(config.ReconciliationKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, errors.Wrap(err, "cannot get concurrent reconcile")
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: className,
		}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: int(concurrentReconcile),
		}), nil
}

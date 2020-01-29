package harbor

import (
	"context"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"

	"github.com/ovh/harbor-operator/controllers/harbor"
)

const (
	ConfigPrefix      = "harbor-controller"
	ReconciliationKey = ConfigPrefix + "-max-reconcile"
	WatchChildrenKey  = ConfigPrefix + "-watch-children"
)

const (
	defaultConcurrentReconcile = 1
	defaultWatchChildren       = true
)

func GetConfig() (*harbor.Config, error) {
	watchChildren, err := configstore.Filter().GetItemValueBool(WatchChildrenKey)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return nil, errors.Wrapf(err, "key %s", WatchChildrenKey)
		}

		watchChildren = defaultWatchChildren
	}

	var concurrentReconciles int

	concurrentReconcilesValue, err := configstore.Filter().GetItemValueInt(ReconciliationKey)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return nil, errors.Wrapf(err, "key %s", ReconciliationKey)
		}

		concurrentReconciles = defaultConcurrentReconcile
	} else {
		concurrentReconciles = int(concurrentReconcilesValue)
	}

	return &harbor.Config{
		ConcurrentReconciles: concurrentReconciles,
		WatchChildren:        watchChildren,
	}, nil
}

func New(ctx context.Context, name, version string) (*harbor.Reconciler, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get configuration")
	}

	return harbor.New(ctx, name, version, config)
}

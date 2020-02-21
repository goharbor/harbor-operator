package controllers

import (
	"context"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"

	"github.com/goharbor/harbor-operator/controllers/harbor"
)

const (
	ConfigPrefix      = "harbor-controller"
	ReconciliationKey = ConfigPrefix + "-max-reconcile"
	WatchChildrenKey  = ConfigPrefix + "-watch-children"
	HarborClassKey    = ConfigPrefix + "-class"
)

const (
	DefaultConcurrentReconcile = 1
	DefaultWatchChildren       = true
	DefaultHarborClass         = ""
)

func getWatchChildrenConfiguration() (bool, error) {
	watchChildren, err := configstore.Filter().GetItemValueBool(WatchChildrenKey)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return false, errors.Wrapf(err, "key %s", WatchChildrenKey)
		}

		watchChildren = DefaultWatchChildren
	}

	return watchChildren, nil
}

func getConcurrentConfiguration() (int, error) {
	concurrentReconciles, err := configstore.Filter().GetItemValueInt(ReconciliationKey)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return 0, errors.Wrapf(err, "key %s", ReconciliationKey)
		}

		concurrentReconciles = DefaultConcurrentReconcile
	}

	return int(concurrentReconciles), nil
}

func getHarborClassConfiguration() (string, error) {
	harborClass, err := configstore.Filter().GetItemValue(HarborClassKey)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return "", errors.Wrapf(err, "key %s", HarborClassKey)
		}

		harborClass = DefaultHarborClass
	}

	return harborClass, nil
}

func GetConfig() (*harbor.Config, error) {
	watchChildren, err := getWatchChildrenConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "fail to get watch children configuration")
	}

	concurrentReconciles, err := getConcurrentConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "fail to get concurrent reconciles configuration")
	}

	className, err := getHarborClassConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "fail to get harbor class configuration")
	}

	return &harbor.Config{
		ConcurrentReconciles: concurrentReconciles,
		WatchChildren:        watchChildren,
		ClassName:            className,
	}, nil
}

func New(ctx context.Context, name, version string) (*harbor.Reconciler, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get configuration")
	}

	return harbor.New(ctx, name, version, config)
}

package config

import (
	"fmt"

	"github.com/ovh/configstore"
	"github.com/pkg/errors"
)

const (
	ReconciliationKey = "max-reconcile"
	WatchChildrenKey  = "watch-children"
	HarborClassKey    = "class"
)

const (
	DefaultConcurrentReconcile = 1
	DefaultWatchChildren       = true
	DefaultHarborClass         = ""
)

type Config struct {
	ClassName            string
	ConcurrentReconciles int
	WatchChildren        bool
}

func getWatchChildrenConfiguration(prefix string) (bool, error) {
	key := fmt.Sprintf("%s-%s", prefix, WatchChildrenKey)

	watchChildren, err := configstore.Filter().GetItemValueBool(key)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return false, errors.Wrapf(err, "key %s", key)
		}

		watchChildren = DefaultWatchChildren
	}

	return watchChildren, nil
}

func getConcurrentConfiguration(prefix string) (int, error) {
	key := fmt.Sprintf("%s-%s", prefix, ReconciliationKey)

	concurrentReconciles, err := configstore.Filter().GetItemValueInt(key)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return 0, errors.Wrapf(err, "key %s", key)
		}

		concurrentReconciles = DefaultConcurrentReconcile
	}

	return int(concurrentReconciles), nil
}

func getHarborClassConfiguration(prefix string) (string, error) {
	key := fmt.Sprintf("%s-%s", prefix, HarborClassKey)

	harborClass, err := configstore.Filter().GetItemValue(key)
	if err != nil {
		_, ok := err.(configstore.ErrItemNotFound)
		if !ok {
			return "", errors.Wrapf(err, "key %s", key)
		}

		harborClass = DefaultHarborClass
	}

	return harborClass, nil
}

func GetConfig(prefix string) (*Config, error) {
	watchChildren, err := getWatchChildrenConfiguration(prefix)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get watch children configuration")
	}

	concurrentReconciles, err := getConcurrentConfiguration(prefix)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get concurrent reconciles configuration")
	}

	className, err := getHarborClassConfiguration(prefix)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get harbor class configuration")
	}

	return &Config{
		ConcurrentReconciles: concurrentReconciles,
		WatchChildren:        watchChildren,
		ClassName:            className,
	}, nil
}

package config

import (
	"errors"
	"fmt"

	"github.com/ovh/configstore"
)

const (
	HarborClassKey            = "classname"
	ReconciliationKey         = "max-concurrent-reconciliation"
	NetworkPoliciesEnabledKey = "network-policies"
)

const (
	DefaultPriority = 5

	DefaultConcurrentReconcile    = 1
	DefaultHarborClass            = ""
	DefaultRegistry               = ""
	DefaultNetworkPoliciesEnabled = false
)

var ErrNotReady = errors.New("configuration not ready")

func NewConfigWithDefaults() *configstore.Store {
	defaultStore := configstore.NewStore()
	defaultStore.InMemory("default-controller").Add(
		configstore.NewItem(ReconciliationKey, fmt.Sprintf("%v", DefaultConcurrentReconcile), DefaultPriority),
		configstore.NewItem(HarborClassKey, DefaultHarborClass, DefaultPriority),
		configstore.NewItem(NetworkPoliciesEnabledKey, fmt.Sprintf("%v", DefaultNetworkPoliciesEnabled), DefaultPriority),
	)

	return defaultStore
}

func GetItem(configStore *configstore.Store, name string, defaultValue string) (configstore.Item, error) {
	item, err := configstore.Filter().
		Store(configStore).
		Slice(name).
		GetFirstItem()
	if IsNotFound(err, name) {
		return configstore.NewItem(name, defaultValue, DefaultPriority), nil
	}

	return item, err
}

func GetString(configStore *configstore.Store, name string, defaultValue string) (string, error) {
	item, err := GetItem(configStore, name, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	return item.Value()
}

func GetBool(configStore *configstore.Store, name string, defaultValue bool) (bool, error) {
	item, err := GetItem(configStore, name, fmt.Sprintf("%v", defaultValue))
	if err != nil {
		return defaultValue, err
	}

	return item.ValueBool()
}

func GetInt(configStore *configstore.Store, name string, defaultValue int) (int, error) {
	item, err := GetItem(configStore, name, fmt.Sprintf("%v", defaultValue))
	if err != nil {
		return defaultValue, err
	}

	v, err := item.ValueInt()

	return int(v), err
}

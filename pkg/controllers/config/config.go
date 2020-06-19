package config

import (
	"fmt"

	"github.com/ovh/configstore"
)

const (
	HarborClassKey    = "classname"
	ReconciliationKey = "max-concurrent-reconciliation"
)

const (
	DefaultPriority = 50

	DefaultConcurrentReconcile = 1
	DefaultHarborClass         = ""
)

func NewConfigWithDefaults() *configstore.Store {
	defaultStore := configstore.NewStore()
	defaultStore.InMemory("default-controller").Add(
		configstore.NewItem(ReconciliationKey, fmt.Sprintf("%v", DefaultConcurrentReconcile), DefaultPriority),
		configstore.NewItem(HarborClassKey, DefaultHarborClass, DefaultPriority),
	)

	return defaultStore
}

package config

import (
	"errors"
	"fmt"

	"github.com/ovh/configstore"
)

const (
	HarborClassKey    = "classname"
	ReconciliationKey = "max-concurrent-reconciliation"
)

const (
	DefaultPriority = 5

	DefaultConcurrentReconcile = 1
	DefaultHarborClass         = ""
	DefaultRegistry            = ""
)

var ErrNotReady = errors.New("configuration not ready")

func NewConfigWithDefaults() *configstore.Store {
	defaultStore := configstore.NewStore()
	defaultStore.InMemory("default-controller").Add(
		configstore.NewItem(ReconciliationKey, fmt.Sprintf("%v", DefaultConcurrentReconcile), DefaultPriority),
		configstore.NewItem(HarborClassKey, DefaultHarborClass, DefaultPriority),
	)

	return defaultStore
}

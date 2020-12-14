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
	DefaultPriority = 50

	DefaultConcurrentReconcile = 1
	DefaultHarborClass         = ""
	DefaultRegistry            = ""
	DefaultHarborVersion       = "v2.1.2"
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

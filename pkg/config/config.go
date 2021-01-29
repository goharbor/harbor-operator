package config

import (
	"errors"
	"fmt"

	"github.com/ovh/configstore"
	corev1 "k8s.io/api/core/v1"
)

const (
	HarborClassKey    = "classname"
	ReconciliationKey = "max-concurrent-reconciliation"
)

const (
	DefaultPriority = 5

	DefaultConcurrentReconcile = 1
	DefaultHarborClass         = ""

	// DefaultImagePullPolicy specifies the policy to image pulls.
	DefaultImagePullPolicy = corev1.PullIfNotPresent
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

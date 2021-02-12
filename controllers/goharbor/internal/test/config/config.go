package config

import (
	"context"
	"path"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
)

const PathToAssets = "../../../config/config/assets"

func New(ctx context.Context, templateKey, fileName string) (*configstore.Store, *configstore.InMemoryProvider) {
	configStore := config.NewConfigWithDefaults()
	provider := configStore.InMemory("test")
	provider.Add(configstore.NewItem(templateKey, path.Join(PathToAssets, fileName), 100))

	return configStore, provider
}

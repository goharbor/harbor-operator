package registryctl

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/config"
)

func (r *Reconciler) GetImage(ctx context.Context) (string, error) {
	image, err := r.ConfigStore.GetItemValue(ConfigImageKey)
	if err != nil {
		if !config.IsNotFound(err, ConfigImageKey) {
			return "", err
		}

		image = DefaultImage
	}

	return image, nil
}

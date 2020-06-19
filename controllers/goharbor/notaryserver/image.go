package notaryserver

import (
	"context"

	"github.com/ovh/configstore"
)

func (r *Reconciler) GetImage(ctx context.Context) (string, error) {
	image, err := r.ConfigStore.GetItemValue(ConfigImageKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return "", err
		}

		image = DefaultImage
	}

	return image, nil
}

func (r *Reconciler) GetDBMigratorImage(ctx context.Context) (string, error) {
	return "migrate/migrate", nil
}

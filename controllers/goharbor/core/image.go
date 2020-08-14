package core

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

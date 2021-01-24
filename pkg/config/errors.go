package config

import (
	"errors"
	"fmt"

	"github.com/ovh/configstore"
)

func IsNotFound(err error, key string) bool {
	// value
	if errors.Is(err, configstore.ErrItemNotFound(fmt.Sprintf("configstore: get '%s': no item found", key))) {
		return true
	}

	// slices
	if errors.Is(err, configstore.ErrItemNotFound(fmt.Sprintf("configstore: get first item (slice: %s): no item found", key))) {
		return true
	}

	return false
}

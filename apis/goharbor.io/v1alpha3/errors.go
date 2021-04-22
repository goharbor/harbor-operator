package v1alpha3

import "errors"

var (
	ErrNoStorageConfiguration = errors.New("no storage configuration")
	Err2StorageConfiguration  = errors.New("only 1 storage can be configured")
)

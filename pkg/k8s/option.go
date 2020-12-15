package k8s

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/ovh/configstore"
	"k8s.io/apimachinery/pkg/runtime"
)

type CtrlOptions struct {
	CTX         context.Context
	Client      Client
	Log         logr.Logger
	DClient     DClient
	Scheme      *runtime.Scheme
	ConfigStore *configstore.Store
}

type Option func(ops *CtrlOptions)

func WithClient(client Client) Option {
	return func(ops *CtrlOptions) {
		ops.Client = client
	}
}

func WithDClient(dClient DClient) Option {
	return func(ops *CtrlOptions) {
		ops.DClient = dClient
	}
}

func WithScheme(scheme *runtime.Scheme) Option {
	return func(ops *CtrlOptions) {
		ops.Scheme = scheme
	}
}

func WithLog(log logr.Logger) Option {
	return func(ops *CtrlOptions) {
		ops.Log = log
	}
}

func WithConfigStore(store *configstore.Store) Option {
	return func(ops *CtrlOptions) {
		ops.ConfigStore = store
	}
}

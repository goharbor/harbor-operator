package k8s

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
)

type GetOptions struct {
	CXT     context.Context
	Client  Client
	Log     logr.Logger
	DClient DClient
	Scheme  *runtime.Scheme
}

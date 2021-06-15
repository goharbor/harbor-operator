package logger

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var loggerContext = "logger"

func Get(ctx context.Context) logr.Logger {
	l := ctx.Value(&loggerContext)
	if l == nil {
		return ctrl.Log
	}

	return l.(logr.Logger)
}

func Set(ctx *context.Context, log logr.Logger) {
	*ctx = context.WithValue(*ctx, &loggerContext, log)
}

func Context(log logr.Logger) context.Context {
	ctx := context.TODO()
	Set(&ctx, log)

	return ctx
}

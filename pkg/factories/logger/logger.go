package logger

import (
	"context"

	"github.com/go-logr/logr"
)

var (
	loggerContext = "logger"
)

func Get(ctx context.Context) logr.Logger {
	return ctx.Value(&loggerContext).(logr.Logger)
}

func Set(ctx *context.Context, log logr.Logger) {
	*ctx = context.WithValue(*ctx, &loggerContext, log)
}

func Context(log logr.Logger) context.Context {
	ctx := context.TODO()
	Set(&ctx, log)

	return ctx
}

package application

import (
	"context"
)

var (
	appNameContext    = "app-name"
	appVersionContext = "app-version"
)

func GetName(ctx context.Context) string {
	return ctx.Value(&appNameContext).(string)
}

func SetName(ctx *context.Context, name string) {
	*ctx = context.WithValue(*ctx, &appNameContext, name)
}

func GetVersion(ctx context.Context) string {
	return ctx.Value(&appVersionContext).(string)
}

func SetVersion(ctx *context.Context, name string) {
	*ctx = context.WithValue(*ctx, &appVersionContext, name)
}

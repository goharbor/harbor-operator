package components

import (
	"context"
)

var (
	componentContext = "component"
	resourceContext  = "resource"
)

func ComponentName(ctx context.Context) string {
	return ctx.Value(&componentContext).(string)
}

func withComponent(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, &componentContext, name)
}

func ResourceName(ctx context.Context) string {
	return ctx.Value(&resourceContext).(string)
}

func withResource(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, &resourceContext, name)
}

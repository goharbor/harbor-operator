package application

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	appNameContext        = "app-name"
	appVersionContext     = "app-version"
	appGitCommitContext   = "app-git-commit"
	appDeletableResources = "app-deletable-resources"
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

func SetVersion(ctx *context.Context, version string) {
	*ctx = context.WithValue(*ctx, &appVersionContext, version)
}

func GetGitCommit(ctx context.Context) string {
	return ctx.Value(&appGitCommitContext).(string)
}

func SetGitCommit(ctx *context.Context, gitCommit string) {
	*ctx = context.WithValue(*ctx, &appGitCommitContext, gitCommit)
}

func GetDeletableResources(ctx context.Context) map[schema.GroupVersionKind]struct{} {
	deletableResources := ctx.Value(&appDeletableResources)
	if deletableResources == nil {
		return nil
	}

	return deletableResources.(map[schema.GroupVersionKind]struct{})
}

func SetDeletableResources(ctx *context.Context, deletableResources map[schema.GroupVersionKind]struct{}) {
	*ctx = context.WithValue(*ctx, &appDeletableResources, deletableResources)
}

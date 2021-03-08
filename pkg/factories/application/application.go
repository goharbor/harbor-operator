package application

import (
	"context"
)

var (
	appNameContext      = "app-name"
	appVersionContext   = "app-version"
	appGitCommitContext = "app-git-commit"
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

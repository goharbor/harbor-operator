package owner

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ownerContext = "owner"
)

func Get(ctx context.Context) metav1.Object {
	owner := ctx.Value(&ownerContext)
	if owner == nil {
		return nil
	}

	return owner.(metav1.Object)
}

func Set(ctx *context.Context, object metav1.Object) {
	*ctx = context.WithValue(*ctx, &ownerContext, object)
}

func Context(object metav1.Object) context.Context {
	ctx := context.TODO()
	Set(&ctx, object)

	return ctx
}

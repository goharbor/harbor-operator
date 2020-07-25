package owner

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var ownerContext = "owner"

type Owner interface {
	runtime.Object
	metav1.Object
}

func Get(ctx context.Context) Owner {
	owner := ctx.Value(&ownerContext)
	if owner == nil {
		return nil
	}

	return owner.(Owner)
}

func Set(ctx *context.Context, object Owner) {
	*ctx = context.WithValue(*ctx, &ownerContext, object)
}

func Context(object Owner) context.Context {
	ctx := context.TODO()
	Set(&ctx, object)

	return ctx
}

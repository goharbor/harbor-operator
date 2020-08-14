package controller

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/opentracing/opentracing-go"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c *Controller) AddGVKToSpan(ctx context.Context, span opentracing.Span, resource runtime.Object) (gvk schema.GroupVersionKind) {
	gvks, _, err := c.Scheme.ObjectKinds(resource)
	if err != nil {
		logger.Get(ctx).Error(err, "cannot get object kind", "resource", resource)

		span.SetTag("Resource.Kind", gvk)
	} else {
		gvk = gvks[0]
	}

	return
}

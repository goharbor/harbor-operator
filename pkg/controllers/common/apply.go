package common

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (c *Controller) Apply(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return errors.Errorf("unsupported resource type %+v", node)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "applyResource", opentracing.Tags{
		"Resource.Kind": res.resource.GetObjectKind().GroupVersionKind().GroupKind(),
	})
	defer span.Finish()

	objectKey, err := client.ObjectKeyFromObject(res.resource)
	if err != nil {
		return errors.Wrap(err, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	result := res.resource.DeepCopyObject()

	op, err := controllerutil.CreateOrUpdate(ctx, c.Client, result, res.mutable(ctx, res.resource, result))
	if err != nil {
		return errors.Wrapf(err, "cannot create/update %+v", res.resource)
	}

	span.SetTag("Operation.Result", op)

	return nil
}

package common

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (c *Controller) Create(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return errors.Errorf("unsupported resource type %+v", node)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "applyResource", opentracing.Tags{})
	defer span.Finish()

	var gvk schema.GroupVersionKind

	gvks, _, err := c.Scheme.ObjectKinds(res.resource)
	if err != nil {
		logger.Get(ctx).Error(err, "cannot get object kind", "resource", res)

		gvk = gvks[0]

		span.SetTag("Resource.Kind", gvk)
	}

	objectKey, err := client.ObjectKeyFromObject(res.resource)
	if err != nil {
		return errors.Wrap(err, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	err = c.Client.Create(ctx, res.resource)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			return nil
		}

		return errors.Wrapf(err, "cannot create %s %s/%s", gvk, res.resource.GetNamespace(), res.resource.GetNamespace())
	}

	return nil
}

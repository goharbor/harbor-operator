package controller

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Controller) Create(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return errors.Errorf("unsupported resource type %+v", node)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "applyResource")
	defer span.Finish()

	gvk := c.AddGVKToSpan(ctx, span, res.Resource)

	objectKey, err := client.ObjectKeyFromObject(res.Resource)
	if err != nil {
		return errors.Wrap(err, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	err = c.Client.Create(ctx, res.Resource)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			return nil
		}

		return errors.Wrapf(err, "cannot create %s %s/%s", gvk, res.Resource.GetNamespace(), res.Resource.GetNamespace())
	}

	return nil
}

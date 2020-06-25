package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

var errNotReady = errors.New("not ready")

func (c *Controller) EnsureReady(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return errors.Errorf("unsupported resource type %+v", node)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "checkReady", opentracing.Tags{})
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
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	result := res.resource.DeepCopyObject()

	err = c.Client.Get(ctx, objectKey, result)
	if err != nil {
		// TODO Check if the error is a temporary error or a unrecoverrable one
		return errors.Wrapf(err, "cannot get %s %s/%s", gvk, res.resource.GetNamespace(), res.resource.GetName())
	}

	ok, err = res.checkable(ctx, result)
	if err != nil {
		return errors.Wrap(err, "cannot check resource status")
	}

	if !ok {
		return serrors.RetryLaterError(errNotReady, "dependencyStatus", fmt.Sprintf("%s %v", result.GetObjectKind().GroupVersionKind().GroupKind(), objectKey), 0*time.Second)
	}

	return nil
}

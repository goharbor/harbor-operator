package common

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serrors "github.com/goharbor/harbor-operator/pkg/controllers/common/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (c *Controller) EnsureReady(ctx context.Context, node graph.Resource) error {
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
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	result := res.resource.DeepCopyObject()

	err = c.Client.Get(ctx, objectKey, result)
	if err != nil {
		// TODO Check if the error is a temporary error or a unrecoverrable one
		return errors.Wrapf(err, "cannot get resource %+v", res.resource)
	}

	ok, err = res.checkable(ctx, result)
	if err != nil {
		return errors.Wrap(err, "cannot check resource status")
	}

	if !ok {
		err := errors.New("not ready")
		return serrors.RetryLaterError(err, "dependencyStatus", fmt.Sprintf("%s %s", result.GetObjectKind().GroupVersionKind().GroupKind(), objectKey), 0*time.Second)
	}

	return nil
}

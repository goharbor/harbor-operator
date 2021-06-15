package controller

import (
	"context"
	"fmt"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errNotReady = errors.New("not ready")

func (c *Controller) ensureResourceReady(ctx context.Context, res *Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "checkReady")
	defer span.Finish()

	objectKey := client.ObjectKeyFromObject(res.resource)

	gvk := c.AddGVKToSpan(ctx, span, res.resource)
	l := logger.Get(ctx)

	result := res.resource.DeepCopyObject().(resources.Resource)

	err := c.Client.Get(ctx, objectKey, result)
	if err != nil {
		// TODO Check if the error is a temporary error or a unrecoverrable one
		return errors.Wrapf(err, "cannot get %s %s/%s", gvk, res.resource.GetNamespace(), res.resource.GetName())
	}

	checksum.CopyMarkers(result.(metav1.Object), res.resource)

	l.V(1).Info("Checking resource readiness")

	ok, err := res.checkable(ctx, result)
	if err != nil {
		return errors.Wrap(err, "cannot check resource status")
	}

	if !ok {
		l.Info("Resource is not ready")

		return serrors.RetryLaterError(errNotReady, "dependencyStatus", fmt.Sprintf("%s %v", result.GetObjectKind().GroupVersionKind().GroupKind(), objectKey))
	}

	l.Info("Resource is ready")

	return nil
}

func (c *Controller) EnsureReady(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", node), serrors.OperatorReason, "unable to apply resource")
	}

	return c.ensureResourceReady(ctx, res)
}

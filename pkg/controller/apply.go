package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

const (
	RetryDuration = 30 * time.Second
	RetryDelay    = time.Second
)

func (c *Controller) apply(ctx context.Context, res *Resource) error {
	retry, ctx := errgroup.WithContext(ctx)
	l := logger.Get(ctx).WithValues("resource.namespace", res.resource.GetNamespace(), "resource.name", res.resource.GetName())

	end := time.Now().Add(RetryDuration)

	var f func() error

	f = func() error {
		span, ctx := opentracing.StartSpanFromContext(ctx, "createOrUpdate", &opentracing.Tags{})
		defer span.Finish()

		gvk := c.AddGVKToSpan(ctx, span, res.resource)

		l.V(1).Info("Deploying resource", "gvk", gvk)

		result := res.resource.DeepCopyObject()

		op, err := controllerutil.CreateOrUpdate(ctx, c.Client, result, res.mutable(ctx, res.resource, result))
		if err != nil {
			span.SetTag("error", err)

			if apierrs.IsConflict(err) {
				if time.Now().After(end) {
					return errors.Wrap(err, "max retry exceeded")
				}

				l.Info(fmt.Sprintf("Failed to update resource, retrying in %v...", RetryDelay), "resource", res.resource)

				time.Sleep(RetryDelay)
				retry.Go(f)

				return nil
			}

			// TODO Check if the error is a temporary error or a unrecoverrable one
			return errors.Wrapf(err, "cannot create/update %s (%s/%s)", gvk, res.resource.GetNamespace(), res.resource.GetName())
		}

		span.SetTag("Operation.Result", op)

		l.Info("Resource deployed", "resource.apiVersion", gvk.GroupVersion(), "resource.kind", gvk.Kind)

		return nil
	}

	retry.Go(f)

	return retry.Wait()
}

func (c *Controller) Apply(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", node), serrors.OperatorReason, "unable to apply resource")
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "applyResource", opentracing.Tags{})
	defer span.Finish()

	if kinds, unversioned, err := c.Scheme.ObjectKinds(res.resource); err == nil {
		span.
			SetTag("Resource.Kind", kinds[0].Kind).
			SetTag("Resource.Versioned", !unversioned)
	} else {
		logger.Get(ctx).Error(err, "Cannot find kinds", "resource", res.resource)
	}

	objectKey, err := client.ObjectKeyFromObject(res.resource)
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get resource key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	return c.apply(ctx, res)
}

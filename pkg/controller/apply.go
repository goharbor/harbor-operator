package controller

import (
	"context"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Controller) Apply(ctx context.Context, res *Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "apply")
	defer span.Finish()

	l := logger.Get(ctx).WithName("resource_applier")

	l.V(1).Info("Deploying resource")

	resource := res.resource

	if err := res.mutable(ctx, resource); err != nil {
		return errors.Wrap(err, "mutate")
	}

	key := client.ObjectKeyFromObject(resource)

	existing := resource.DeepCopyObject().(client.Object)
	if err := c.Get(ctx, key, existing); err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}

		l.Info("apply creating", "key", key, "kind", resource.GetObjectKind().GroupVersionKind())

		if err := c.Create(ctx, resource); err != nil {
			return err
		}

		return nil
	}

	l.Info("apply changing", "key", key, "kind", resource.GetObjectKind().GroupVersionKind())

	if err := c.Client.Update(ctx, resource, &client.UpdateOptions{
		FieldManager: application.GetName(ctx),
	}); err != nil {
		l.Error(err, "Cannot deploy resource", "key", key, "kind", resource.GetObjectKind().GroupVersionKind())

		if apierrs.IsForbidden(err) {
			return serrors.RetryLaterError(err, "dependencyStatus", err.Error())
		}

		if apierrs.IsInvalid(err) {
			return serrors.UnrecoverrableError(err, "dependencySpec", err.Error())
		}

		return err
	}

	return nil
}

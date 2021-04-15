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

var force = true

func (c *Controller) Apply(ctx context.Context, res *Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "apply")
	defer span.Finish()

	l := logger.Get(ctx)

	l.V(1).Info("Deploying resource")

	resource := res.resource

	if err := res.mutable(ctx, resource); err != nil {
		return errors.Wrap(err, "mutate")
	}

	err := c.Client.Patch(ctx, resource, client.Apply, &client.PatchOptions{
		Force:        &force,
		FieldManager: application.GetName(ctx),
	})
	if err != nil {
		l.Error(err, "Cannot deploy resource")

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

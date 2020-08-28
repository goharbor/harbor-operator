package controller

import (
	"context"
	"fmt"
	"time"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	RetryDuration = 30 * time.Second
	RetryDelay    = time.Second
)

func (c *Controller) apply(ctx context.Context, res *Resource) (controllerutil.OperationResult, error) {
	retry, ctx := errgroup.WithContext(ctx)
	l := logger.Get(ctx)
	end := time.Now().Add(RetryDuration)
	opResult := controllerutil.OperationResultNone

	var f func() error

	f = func() error {
		span, ctx := opentracing.StartSpanFromContext(ctx, "applyAndRetry")
		defer span.Finish()

		result := res.resource.DeepCopyObject()

		op, err := controllerutil.CreateOrUpdate(ctx, c.Client, result, res.mutable(ctx, res.resource, result))
		if err != nil { //nolint:nestif
			span.SetTag("error", err)

			if apierrs.IsConflict(err) {
				if time.Now().After(end) {
					return errors.Wrap(err, "max retry exceeded")
				}

				l.V(1).Info(fmt.Sprintf("Failed to update resource, retrying in %v...", RetryDelay))

				time.Sleep(RetryDelay)
				retry.Go(f)

				return nil
			}

			if apierrs.IsForbidden(err) {
				return serrors.RetryLaterError(err, "dependencyStatus", err.Error())
			}

			if apierrs.IsInvalid(err) {
				return serrors.UnrecoverrableError(err, "dependencySpec", err.Error())
			}

			// TODO Check if the error is a temporary error or a unrecoverrable one
			return errors.Wrapf(err, "cannot create/update")
		}

		span.SetTag("operation.result", op)

		checksum.CopyMarkers(result.(metav1.Object), res.resource)

		opResult = op

		return nil
	}

	retry.Go(f)

	err := retry.Wait()

	return opResult, err
}

func (c *Controller) Apply(ctx context.Context, res *Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "apply")
	defer span.Finish()

	l := logger.Get(ctx)

	l.V(1).Info("Deploying resource")

	op, err := c.apply(ctx, res)
	if err != nil {
		l.Error(err, "Cannot deploy resource")

		return err
	}

	l.Info("Resource deployed", "operation.result", op)

	return nil
}

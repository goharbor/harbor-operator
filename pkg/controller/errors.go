package controller

import (
	"context"
	"strings"
	"time"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

type causer interface {
	Cause() error
}

func IsOptimisticLockError(err error) bool {
	return strings.Contains(err.Error(), optimisticLockErrorMsg)
}

func (c *Controller) HandleError(ctx context.Context, resource client.Object, resultError error) (ctrl.Result, error) {
	if resultError == nil {
		return ctrl.Result{}, c.SetSuccessStatus(ctx, resource)
	}

	// Do manual retry without error when resultError is an optimistic lock error.
	// For more info, see https://github.com/kubernetes/kubernetes/issues/28149
	if IsOptimisticLockError(resultError) {
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "handleError", opentracing.Tags{
		"Resource": resource,
		"error":    resultError,
	})
	defer span.Finish()

	objectKey := client.ObjectKeyFromObject(resource)

	err := c.Client.Get(ctx, objectKey, resource)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(resultError, errors.Wrap(err, "cannot get object").Error())
	}

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(resultError, errors.Wrap(err, "cannot convert resource to unstuctured").Error())
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	err = c.SetErrorStatus(ctx, u, resultError)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(resultError, errors.Wrap(err, "cannot set status to error").Error())
	}

	logger.Get(ctx).Info("error reported to resource status", "error", resultError.Error())

	return ctrl.Result{}, nil
}

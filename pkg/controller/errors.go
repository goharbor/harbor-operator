package controller

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type causer interface {
	Cause() error
}

func (c *Controller) HandleError(ctx context.Context, resource runtime.Object, resultError error) (ctrl.Result, error) {
	if resultError == nil {
		return ctrl.Result{}, c.SetSuccessStatus(ctx, resource)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "handleError", opentracing.Tags{
		"Resource": resource,
		"error":    resultError,
	})
	defer span.Finish()

	objectKey, err := client.ObjectKeyFromObject(resource)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(resultError, errors.Wrap(err, "cannot get object key").Error())
	}

	err = c.Client.Get(ctx, objectKey, resource)
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

	logger.Get(ctx).Info("error handled")

	return ctrl.Result{}, nil
}

package common

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"

	serrors "github.com/goharbor/harbor-operator/pkg/controllers/common/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	sstatus "github.com/goharbor/harbor-operator/pkg/status"
)

type causer interface {
	Cause() error
}

func (c *Controller) HandleError(ctx context.Context, resource runtime.Object, resultError error) (ctrl.Result, error) {
	if resultError == nil {
		return ctrl.Result{}, nil
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

	result, resultError := postUpdateData(ctx, u, resultError)

	err = c.Client.Status().Update(ctx, u)
	if err != nil {
		return result, errors.Wrap(resultError, errors.Wrap(err, "cannot update status").Error())
	}

	if resultError != nil {
		return result, resultError
	}

	logger.Get(ctx).Info("error handled")

	return result, nil
}

func postUpdateData(ctx context.Context, u *unstructured.Unstructured, resultError error) (ctrl.Result, error) {
	result := ctrl.Result{}

	err := status.Augment(u)
	if err != nil {
		return result, errors.Wrap(resultError, errors.Wrap(err, "unable to augment resource status").Error())
	}

	data := u.UnstructuredContent()

	conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
	if err != nil {
		return result, errors.Wrap(resultError, errors.Wrap(err, "cannot get conditions").Error())
	}

	errLoop := resultError
	for errLoop != nil {
		stop, conds, r, re := getStatusResult(resultError, errLoop)
		result, resultError = r, re

		for _, cond := range conds {
			conds, err := sstatus.UpdateCondition(ctx, conditions, cond.Type, cond.Status, cond.Reason, cond.Message)
			if err != nil {
				return result, errors.Wrap(resultError, errors.Wrapf(err, "cannot update %s condition to %s", cond.Type, cond.Status).Error())
			}

			conditions = conds
		}

		if stop {
			break
		}

		cause, ok := errLoop.(causer)
		if !ok {
			break
		}

		errLoop = cause.Cause()
	}

	conditions, err = sstatus.UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionTrue, "recoverrableError", "An error occurred and may be recovered")
	if err != nil {
		return result, errors.Wrap(resultError, errors.Wrapf(err, "cannot update %s condition to %s", status.ConditionInProgress, corev1.ConditionTrue).Error())
	}

	err = unstructured.SetNestedSlice(data, conditions, "status", "conditions")
	if err != nil {
		return result, errors.Wrap(resultError, errors.Wrap(err, "cannot update conditions").Error())
	}

	u.SetUnstructuredContent(data)

	return result, resultError
}

func getStatusResult(rootError, localErr error) (stop bool, conditions []status.Condition, result ctrl.Result, resultError error) {
	stop = false
	resultError = rootError

	if err, ok := localErr.(serrors.Resulter); ok {
		result, resultError = err.Result()
		stop = true
	}

	if err, ok := localErr.(serrors.Stature); ok {
		conditions = err.Status()
		stop = true
	}

	return stop, conditions, result, resultError
}

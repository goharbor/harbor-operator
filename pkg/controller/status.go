package controller

import (
	"context"
	"fmt"

	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/resources"
	sstatus "github.com/goharbor/harbor-operator/pkg/status"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func (c *Controller) prepareStatus(ctx context.Context, owner resources.Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prepareStatus")
	defer span.Finish()

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(owner)
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to convert resource to unstuctured")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	stop, err := c.preUpdateData(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot update observedGeneration")
	}

	if stop {
		logger.Get(ctx).Info("nothing to do")
		return nil
	}

	err = c.Client.Status().Update(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot update status")
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), owner)

	return errors.Wrap(err, "cannot update owner")
}

func (c *Controller) SetSuccessStatus(ctx context.Context, resource runtime.Object) error {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return errors.Wrap(err, "cannot convert resource to unstuctured")
	}

	err = c.SetControllerStatus(ctx, data)
	if err != nil {
		return errors.Wrap(err, "cannot set controller status")
	}

	err = c.SetSuccessConditions(ctx, data)
	if err != nil {
		return errors.Wrap(err, "cannot set conditions to success")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	err = c.Client.Status().Update(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot update status")
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), resource)

	return errors.Wrap(err, "cannot update resource")
}

func (c *Controller) SetSuccessConditions(ctx context.Context, data map[string]interface{}) error {
	conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
	if err != nil {
		return errors.Wrap(err, "cannot get conditions")
	}

	conditions, err = sstatus.UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionFalse)
	if err != nil {
		return errors.Wrapf(err, "cannot update %s condition to %s", status.ConditionInProgress, corev1.ConditionFalse)
	}

	conditions, err = sstatus.UpdateCondition(ctx, conditions, status.ConditionFailed, corev1.ConditionFalse)
	if err != nil {
		return errors.Wrapf(err, "cannot update %s condition to %s", status.ConditionFailed, corev1.ConditionFalse)
	}

	err = unstructured.SetNestedSlice(data, conditions, "status", "conditions")

	return errors.Wrap(err, "cannot update conditions")
}

func (c *Controller) preUpdateData(ctx context.Context, u *unstructured.Unstructured) (bool, error) {
	err := status.Augment(u)
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to augment resource status")
	}

	data := u.UnstructuredContent()

	stopByControllerVersion, err := c.preUpdateControllerStatus(ctx, data)
	if err != nil {
		return false, err
	}

	stopByGeneration, err := c.preUpdateObservedGeneration(ctx, data)
	if err != nil {
		return false, err
	}

	u.SetUnstructuredContent(data)

	return stopByControllerVersion && stopByGeneration, nil
}

func (c *Controller) preUpdateControllerStatus(ctx context.Context, data map[string]interface{}) (bool, error) {
	observedVersion, foundVersion, err := unstructured.NestedString(data, "status", "operator", "controllerVersion")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get controller version")
	}

	observedName, foundName, err := unstructured.NestedString(data, "status", "operator", "controllerName")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get controller name")
	}

	err = c.SetControllerStatus(ctx, data)
	if err != nil {
		return false, errors.Wrap(err, "cannot set controller status")
	}

	version, name := c.GetVersion(), c.GetName()

	logger.Get(ctx).V(1).Info(
		"Updating controller status",
		"oldVersion", observedVersion,
		"newVersion", version,
		"oldName", observedName,
		"newName", name,
	)

	return foundVersion && foundName && version == observedVersion && name == observedName, nil
}

func (c *Controller) SetControllerStatus(ctx context.Context, data map[string]interface{}) error {
	err := unstructured.SetNestedField(data, c.GetVersion(), "status", "operator", "controllerVersion")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to set observed generation")
	}

	err = unstructured.SetNestedField(data, c.GetName(), "status", "operator", "controllerName")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to set observed generation")
	}

	return nil
}

var errNoGeneration = errors.New("no $.metadata.generation found")

func (c *Controller) preUpdateObservedGeneration(ctx context.Context, data map[string]interface{}) (bool, error) {
	generation, found, err := unstructured.NestedInt64(data, "metadata", "generation")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get generation")
	}

	if !found {
		return false, serrors.UnrecoverrableError(errNoGeneration, serrors.OperatorReason, "unable to get generation")
	}

	observedGeneration, found, err := unstructured.NestedInt64(data, "status", "observedGeneration")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get observed generation")
	}

	err = unstructured.SetNestedField(data, generation, "status", "observedGeneration")
	if err != nil {
		return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to set observed generation")
	}

	// Already observed
	if found && generation == observedGeneration {
		conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
		if err != nil {
			return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get conditions")
		}

		failedStatus, err := sstatus.GetConditionStatus(ctx, conditions, status.ConditionFailed)
		if err != nil {
			return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, fmt.Sprintf("unable to check %s condition", status.ConditionFailed))
		}

		inProgressStatus, err := sstatus.GetConditionStatus(ctx, conditions, status.ConditionInProgress)
		if err != nil {
			return false, serrors.UnrecoverrableError(err, serrors.OperatorReason, fmt.Sprintf("unable to check %s condition", status.ConditionInProgress))
		}

		return inProgressStatus == corev1.ConditionFalse && failedStatus == corev1.ConditionFalse, nil
	}

	return false, c.preUpdateConditions(ctx, data)
}

func (c *Controller) preUpdateConditions(ctx context.Context, data map[string]interface{}) error {
	conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get conditions")
	}

	conditions, err = sstatus.UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionTrue, "newGeneration", "New generation detected")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to update condition")
	}

	err = unstructured.SetNestedSlice(data, conditions, "status", "conditions")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to update condition")
	}

	return nil
}

func (c *Controller) SetErrorStatus(ctx context.Context, resource runtime.Object, resultError error) error {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return errors.Wrap(err, "cannot convert resource to unstuctured")
	}

	err = c.SetControllerStatus(ctx, data)
	if err != nil {
		return errors.Wrap(err, "cannot set controller status")
	}

	err = c.SetErrorConditions(ctx, data, resultError)
	if err != nil {
		return errors.Wrap(err, "cannot set conditions to error")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	err = c.Client.Status().Update(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot update status")
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), resource)

	return errors.Wrap(err, "cannot update resource")
}

func (c *Controller) SetErrorConditions(ctx context.Context, data map[string]interface{}, resultError error) error {
	conditions, _, err := unstructured.NestedSlice(data, "status", "conditions")
	if err != nil {
		return errors.Wrap(resultError, errors.Wrap(err, "cannot get conditions").Error())
	}

	errLoop := resultError
	for errLoop != nil {
		stop, conds, re := c.getStatusResult(resultError, errLoop)
		resultError = re

		for _, cond := range conds {
			conds, err := sstatus.UpdateCondition(ctx, conditions, cond.Type, cond.Status, cond.Reason, cond.Message)
			if err != nil {
				return errors.Wrap(resultError, errors.Wrapf(err, "cannot update %s condition to %s", cond.Type, cond.Status).Error())
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
		return errors.Wrap(resultError, errors.Wrapf(err, "cannot update %s condition to %s", status.ConditionInProgress, corev1.ConditionTrue).Error())
	}

	err = unstructured.SetNestedSlice(data, conditions, "status", "conditions")
	if err != nil {
		return errors.Wrap(resultError, errors.Wrap(err, "cannot update conditions").Error())
	}

	return resultError
}

func (c *Controller) getStatusResult(rootError, localErr error) (stop bool, conditions []status.Condition, resultError error) {
	stop = false
	resultError = rootError

	if err, ok := localErr.(serrors.Resulter); ok {
		_, resultError = err.Result()
		stop = true
	}

	if err, ok := localErr.(serrors.Status); ok {
		conditions = err.Status()
		stop = true
	}

	return stop, conditions, resultError
}

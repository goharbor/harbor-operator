package controller

import (
	"context"

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

func (c *Controller) PrepareStatus(ctx context.Context, owner resources.Resource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prepareStatus")
	defer span.Finish()

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(owner)
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to convert resource to unstuctured")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	if err := c.preUpdateData(ctx, u); err != nil {
		return errors.Wrap(err, "cannot update observedGeneration")
	}

	if err := c.Client.Status().Update(ctx, u); err != nil {
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

	if err := c.SetControllerStatus(ctx, data); err != nil {
		return errors.Wrap(err, "cannot set controller status")
	}

	if err := c.SetSuccessConditions(ctx, data); err != nil {
		return errors.Wrap(err, "cannot set conditions to success")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	if err := c.Client.Status().Update(ctx, u); err != nil {
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

func (c *Controller) preUpdateData(ctx context.Context, u *unstructured.Unstructured) error {
	if err := status.Augment(u); err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to augment resource status")
	}

	data := u.UnstructuredContent()

	if err := c.preUpdateControllerStatus(ctx, data); err != nil {
		return err
	}

	if err := c.preUpdateObservedGeneration(ctx, data); err != nil {
		return err
	}

	u.SetUnstructuredContent(data)

	return nil
}

func (c *Controller) preUpdateControllerStatus(ctx context.Context, data map[string]interface{}) error {
	observedGitCommit, _, err := unstructured.NestedString(data, "status", "operator", "controllerGitCommit")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get controller git commit")
	}

	observedVersion, _, err := unstructured.NestedString(data, "status", "operator", "controllerVersion")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get controller version")
	}

	observedName, _, err := unstructured.NestedString(data, "status", "operator", "controllerName")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get controller name")
	}

	if err := c.SetControllerStatus(ctx, data); err != nil {
		return errors.Wrap(err, "cannot set controller status")
	}

	gitCommit, version, name := c.GetGitCommit(), c.GetVersion(), c.GetName()

	logger.Get(ctx).V(1).Info(
		"Updating controller status",
		"oldGitCommit", observedGitCommit,
		"newGitCommit", gitCommit,
		"oldVersion", observedVersion,
		"newVersion", version,
		"oldName", observedName,
		"newName", name,
	)

	return nil
}

func (c *Controller) SetControllerStatus(ctx context.Context, data map[string]interface{}) error {
	err := unstructured.SetNestedField(data, c.GetGitCommit(), "status", "operator", "controllerGitCommit")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to set observed generation")
	}

	err = unstructured.SetNestedField(data, c.GetVersion(), "status", "operator", "controllerVersion")
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

func (c *Controller) preUpdateObservedGeneration(ctx context.Context, data map[string]interface{}) error {
	generation, found, err := unstructured.NestedInt64(data, "metadata", "generation")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get generation")
	}

	if !found {
		return serrors.UnrecoverrableError(errNoGeneration, serrors.OperatorReason, "unable to get generation")
	}

	observedGeneration, found, err := unstructured.NestedInt64(data, "status", "observedGeneration")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to get observed generation")
	}

	err = unstructured.SetNestedField(data, generation, "status", "observedGeneration")
	if err != nil {
		return serrors.UnrecoverrableError(err, serrors.OperatorReason, "unable to set observed generation")
	}

	if !found || generation != observedGeneration {
		return c.preUpdateConditions(ctx, data)
	}

	return nil
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

		var cause causer

		if !errors.As(errLoop, &cause) {
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

	var resultErr serrors.Resulter

	if errors.As(localErr, &resultErr) {
		_, resultError = resultErr.Result()
		stop = true
	}

	var statusErr serrors.Status

	if errors.As(localErr, &statusErr) {
		conditions = statusErr.Status()
		stop = true
	}

	return stop, conditions, resultError
}

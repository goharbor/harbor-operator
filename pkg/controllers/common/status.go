package common

import (
	"context"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/kustomize/kstatus/status"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

func (r *Controller) GetCondition(ctx context.Context, componentStatus *goharborv1alpha2.ComponentStatus, conditionType status.ConditionType) goharborv1alpha2.Condition {
	for _, condition := range componentStatus.Conditions {
		if condition.Type == conditionType {
			return condition
		}
	}

	return goharborv1alpha2.Condition{
		Type:   conditionType,
		Status: corev1.ConditionUnknown,
	}
}

func (r *Controller) GetConditionStatus(ctx context.Context, componentStatus *goharborv1alpha2.ComponentStatus, conditionType status.ConditionType) corev1.ConditionStatus {
	return r.GetCondition(ctx, componentStatus, conditionType).Status
}

func (r *Controller) UpdateCondition(ctx context.Context, componentStatus *goharborv1alpha2.ComponentStatus, conditionType status.ConditionType, conditionStatus corev1.ConditionStatus, reasons ...string) error {
	var reason, message string

	switch len(reasons) {
	case 0: // nolint:mnd
	case 1: // nolint:mnd
		reason = reasons[0]
	case 2: // nolint:mnd
		reason = reasons[0]
		message = reasons[1]
	default:
		return errors.Errorf("expecting reason and message, got %d parameters", len(reasons))
	}

	for i, condition := range componentStatus.Conditions {
		if condition.Type == conditionType {
			condition.Status = conditionStatus
			condition.Reason = reason
			condition.Message = message

			componentStatus.Conditions[i] = condition

			return nil
		}
	}

	condition := goharborv1alpha2.Condition{
		Type:    conditionType,
		Status:  conditionStatus,
		Reason:  reason,
		Message: message,
	}

	componentStatus.Conditions = append(componentStatus.Conditions, condition)

	return nil
}

// UpdateStatus applies current in-memory statuses to the remote resource
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#status-subresource
func (r *Controller) UpdateStatus(ctx context.Context, result *ctrl.Result, object runtime.Object) error {
	err := r.Client.Status().Update(ctx, object)
	if err != nil {
		result.Requeue = true

		seconds, needWait := apierrors.SuggestsClientDelay(err)
		if needWait {
			result.RequeueAfter = time.Second * time.Duration(seconds)
		}

		if apierrors.IsConflict(err) {
			// the object has been modified; please apply your changes to the latest version and try again
			logger.Get(ctx).Error(err, "cannot update status field")
			return nil
		}

		return errors.Wrap(err, "cannot update status field")
	}

	return nil
}

package common

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func GetCondition(ctx context.Context, conditions []interface{}, conditionType status.ConditionType) (map[string]interface{}, error) {
	conditionTypeString := conditionType.String()

	for _, condition := range conditions {
		condition, ok := condition.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("invalid condition")
		}

		condType, ok := condition["type"]
		if ok && condType == conditionTypeString {
			return condition, nil
		}
	}

	return map[string]interface{}{
		"type":   conditionType,
		"status": string(corev1.ConditionUnknown),
	}, nil
}

func GetConditionStatus(ctx context.Context, conditions []interface{}, conditionType status.ConditionType) (corev1.ConditionStatus, error) {
	condition, err := GetCondition(ctx, conditions, conditionType)
	if err != nil {
		return "", errors.Wrap(err, "cannot get condition")
	}

	s, ok := condition["status"]
	if !ok {
		return corev1.ConditionUnknown, nil
	}

	result, ok := s.(string)
	if !ok {
		return "", errors.Errorf("invalid status type")
	}

	return corev1.ConditionStatus(result), nil
}

func UpdateCondition(ctx context.Context, conditions []interface{}, conditionType fmt.Stringer, conditionStatus corev1.ConditionStatus, reasons ...string) ([]interface{}, error) {
	var reason, message string

	switch len(reasons) {
	case 0:
	case 1:
		reason = reasons[0]
	case 2: //nolint:gomnd
		reason = reasons[0]
		message = reasons[1]
	default:
		return nil, errors.Errorf("expecting reason and message, got %d parameters", len(reasons))
	}

	conditionTypeString := conditionType.String()

	for _, condition := range conditions {
		condition, ok := condition.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("invalid condition type")
		}

		if condition["type"] == conditionTypeString {
			condition["status"] = string(conditionStatus)

			if reason == "" {
				delete(condition, "reason")
			} else {
				condition["reason"] = reason
			}

			if message == "" {
				delete(condition, "message")
			} else {
				condition["message"] = message
			}

			return conditions, nil
		}
	}

	condition := map[string]interface{}{
		"type":    conditionType.String(),
		"status":  string(conditionStatus),
		"reason":  reason,
		"message": message,
	}

	return append(conditions, condition), nil
}

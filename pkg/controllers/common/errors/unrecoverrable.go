package errors

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

type UnrecoverrableError struct {
	Cause error

	Reason  string
	Message string
}

func (err UnrecoverrableError) Error() string {
	return err.Cause.Error()
}

func (err UnrecoverrableError) Status() status.Condition {
	return status.Condition{
		Type:    status.ConditionFailed,
		Message: err.Message,
		Reason:  err.Message,
		Status:  v1.ConditionTrue,
	}
}

package errors

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

const (
	OperatorReason    = "operatorError"
	InvalidSpecReason = "invalidSpec"
)

type unrecoverrableError struct {
	cause error

	reason  string
	message string
}

func UnrecoverrableError(err error, reason, message string) error {
	return &unrecoverrableError{
		cause:   err,
		reason:  reason,
		message: message,
	}
}

func (err *unrecoverrableError) Error() string {
	return errors.Wrap(err.cause, err.message).Error()
}

func (err *unrecoverrableError) Status() []status.Condition {
	return []status.Condition{
		{
			Type:    status.ConditionInProgress,
			Message: "",
			Reason:  "",
			Status:  v1.ConditionFalse,
		}, {
			Type:    status.ConditionFailed,
			Message: err.cause.Error(),
			Reason:  err.reason,
			Status:  v1.ConditionTrue,
		},
	}
}

func (err *unrecoverrableError) Cause() error {
	return err.cause
}

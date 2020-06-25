package errors

import (
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/kustomize/kstatus/status"
)

const (
	APITemporaryError = "apiTemporaryError"
)

type retryLaterError struct {
	cause error

	delay   time.Duration
	reason  string
	message string
}

func RetryLaterError(err error, reason, message string, delay time.Duration) error {
	return &retryLaterError{
		cause:   err,
		delay:   delay,
		reason:  reason,
		message: message,
	}
}

func (err *retryLaterError) Result() (ctrl.Result, error) {
	if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: err.delay,
		}, nil
	}

	return ctrl.Result{
		Requeue: false,
	}, nil
}

func (err *retryLaterError) Status() []status.Condition {
	return []status.Condition{
		{
			Type:    status.ConditionInProgress,
			Message: "",
			Reason:  "",
			Status:  v1.ConditionTrue,
		}, {
			Type:    status.ConditionFailed,
			Message: err.message,
			Reason:  err.reason,
			Status:  v1.ConditionTrue,
		},
	}
}

func (err *retryLaterError) Error() string {
	return errors.Wrap(err.cause, err.message).Error()
}

func (err *retryLaterError) Cause() error {
	return err.cause
}

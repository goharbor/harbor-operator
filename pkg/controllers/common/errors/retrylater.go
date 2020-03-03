package errors

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type RetryLaterError struct {
	Cause error

	Delay time.Duration
}

func (err *RetryLaterError) Result() ctrl.Result {
	if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: err.Delay,
		}
	}

	return ctrl.Result{
		Requeue: false,
	}
}

func (err RetryLaterError) Error() string {
	return err.Cause.Error()
}

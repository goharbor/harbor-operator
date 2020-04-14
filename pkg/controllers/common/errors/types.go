package errors

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/kustomize/kstatus/status"
)

type Resulter interface {
	Result() (ctrl.Result, error)
}

type Stature interface {
	Status() []status.Condition
}

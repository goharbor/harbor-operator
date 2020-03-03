package resources

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Resource interface {
	runtime.Object
	metav1.Object
}

type Checkable func(context.Context, runtime.Object) (bool, error)

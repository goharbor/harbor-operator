package resources

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resource interface {
	runtime.Object
	metav1.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}

type Checkable func(context.Context, runtime.Object) (bool, error)

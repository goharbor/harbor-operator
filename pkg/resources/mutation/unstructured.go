package mutation

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func NewUnstructured(mutate resources.Mutable) resources.Mutable {
	return func(ctx context.Context, unstructuredResource, unstructuredResult runtime.Object) controllerutil.MutateFn {
		desired := unstructuredResource.(*unstructured.Unstructured)
		result := unstructuredResult.(*unstructured.Unstructured)

		mutate := mutate(ctx, desired, result)

		return func() error {
			result.SetUnstructuredContent(desired.UnstructuredContent())

			return mutate()
		}
	}
}

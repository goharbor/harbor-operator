package mutation

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func NewUnstructured(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, unstructuredResource, unstructuredResult runtime.Object) controllerutil.MutateFn {
		desired := unstructuredResource.(*unstructured.Unstructured)
		result := unstructuredResult.(*unstructured.Unstructured)

		mutate := mutate(ctx, desired, result)

		return func() error {
			gvk := result.GetObjectKind().GroupVersionKind()
			resourceVersion := result.GetResourceVersion()

			result.SetUnstructuredContent(desired.UnstructuredContent())
			result.SetGroupVersionKind(gvk)
			result.SetResourceVersion(resourceVersion)

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}

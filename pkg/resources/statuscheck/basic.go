package statuscheck

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func BasicCheck(ctx context.Context, object runtime.Object) (bool, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return false, errors.Wrap(err, "cannot transform to unstructured")
	}

	resource := &unstructured.Unstructured{}
	resource.SetUnstructuredContent(data)

	return UnstructuredCheck(ctx, resource)
}

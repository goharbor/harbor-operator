package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func ConvertViaUnstructuredCopy(src, dest runtime.Object) error {
	gvk := dest.GetObjectKind().GroupVersionKind()

	defer func() {
		dest.GetObjectKind().SetGroupVersionKind(gvk)
	}()

	srcRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(src)
	if err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(srcRaw, dest)
}

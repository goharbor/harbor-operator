package convert

import "k8s.io/apimachinery/pkg/runtime"

func ConverterObject(o runtime.Object) Converter {
	return convert{Object: o}
}

type Converter interface {
	To(runtime.Object) error
	From(runtime.Object) error
}

type convert struct {
	Object runtime.Object
}

func (o convert) To(dst runtime.Object) error {
	return unstructuredConversion(o.Object, dst)
}

func (o convert) From(src runtime.Object) error {
	return unstructuredConversion(src, o.Object)
}

func unstructuredConversion(src, dest runtime.Object) error {
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

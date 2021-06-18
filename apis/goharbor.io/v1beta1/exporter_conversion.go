package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (e *Exporter) ConvertTo(dstRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(e)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, dstRaw); err != nil {
		return err
	}

	dstRaw.(*v1alpha3.Exporter).APIVersion = v1alpha3.GroupVersion.String()

	return nil
}

func (e *Exporter) ConvertFrom(srcRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(srcRaw)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, e); err != nil {
		return err
	}

	e.APIVersion = GroupVersion.String()

	return nil
}

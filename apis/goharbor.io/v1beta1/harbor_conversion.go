package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Harbor{}

func (h *Harbor) ConvertTo(dstRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(h)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, dstRaw); err != nil {
		return err
	}

	dstRaw.(*v1alpha3.Harbor).APIVersion = v1alpha3.GroupVersion.String()

	return nil
}

func (h *Harbor) ConvertFrom(srcRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(srcRaw)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, h); err != nil {
		return err
	}

	h.APIVersion = GroupVersion.String()

	return nil
}

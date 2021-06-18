package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Harbor{}

func (h *Harbor) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(h, dstRaw)
}

func (h *Harbor) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, h)
}

package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Harbor{}

func (h *Harbor) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(h).To(dstRaw)
}

func (h *Harbor) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(h).From(srcRaw)
}

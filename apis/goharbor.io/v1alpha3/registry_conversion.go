package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Registry{}

func (r *Registry) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(r).To(dstRaw)
}

func (r *Registry) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(r).From(srcRaw)
}

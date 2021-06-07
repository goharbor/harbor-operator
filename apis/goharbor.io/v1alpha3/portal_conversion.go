package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Portal{}

func (p *Portal) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(p).To(dstRaw)
}

func (p *Portal) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(p).From(srcRaw)
}

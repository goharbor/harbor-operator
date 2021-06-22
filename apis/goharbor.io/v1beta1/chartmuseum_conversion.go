package v1beta1

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (c *ChartMuseum) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(c).To(dstRaw)
}

func (c *ChartMuseum) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(c).From(srcRaw)
}

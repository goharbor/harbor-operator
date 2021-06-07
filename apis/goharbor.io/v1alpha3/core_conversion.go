package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Core{}

func (c *Core) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(c).To(dstRaw)
}

func (c *Core) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(c).From(srcRaw)
}

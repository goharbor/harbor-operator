package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (c *ChartMuseum) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(c, dstRaw)
}

func (c *ChartMuseum) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, c)
}

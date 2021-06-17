package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (c *ChartMuseum) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.ChartMuseum)

	return CopyViaJSON(dst, c)
}

func (c *ChartMuseum) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.ChartMuseum)

	return CopyViaJSON(c, src)
}

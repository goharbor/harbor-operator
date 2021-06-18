package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (p *Portal) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(p, dstRaw)
}

func (p *Portal) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, p)
}

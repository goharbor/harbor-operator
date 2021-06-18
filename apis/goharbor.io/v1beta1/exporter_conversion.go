package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (e *Exporter) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(e, dstRaw)
}

func (e *Exporter) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, e)
}

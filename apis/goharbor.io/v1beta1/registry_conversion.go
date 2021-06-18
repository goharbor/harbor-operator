package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (r *Registry) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(r, dstRaw)
}

func (r *Registry) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, r)
}

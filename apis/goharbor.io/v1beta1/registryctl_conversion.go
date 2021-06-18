package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (r *RegistryController) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(r, dstRaw)
}

func (r *RegistryController) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, r)
}

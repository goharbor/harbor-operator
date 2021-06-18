package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (t *Trivy) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(t, dstRaw)
}

func (t *Trivy) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, t)
}

package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (n *NotaryServer) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(n, dstRaw)
}

func (n *NotaryServer) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, n)
}

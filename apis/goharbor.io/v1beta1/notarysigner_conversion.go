package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (n *NotarySigner) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(n, dstRaw)
}

func (n *NotarySigner) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, n)
}

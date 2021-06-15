package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Harbor{}

func (h *Harbor) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Harbor)

	return CopyViaJSON(dst, h)
}

func (h *Harbor) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Harbor)

	return CopyViaJSON(h, src)
}

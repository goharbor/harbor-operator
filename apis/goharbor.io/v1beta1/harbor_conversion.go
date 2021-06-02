package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Harbor{}

func (src *Harbor) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Harbor)
	return CopyViaJSON(dst, src)
}

func (dst *Harbor) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Harbor)
	return CopyViaJSON(dst, src)
}

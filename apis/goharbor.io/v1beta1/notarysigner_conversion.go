package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *NotarySigner) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.NotarySigner)
	return CopyViaJSON(dst, src)
}

func (dst *NotarySigner) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.NotarySigner)
	return CopyViaJSON(dst, src)
}

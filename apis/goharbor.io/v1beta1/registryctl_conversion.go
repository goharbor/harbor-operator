package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RegistryController) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.RegistryController)
	return CopyViaJSON(dst, src)
}

func (dst *RegistryController) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.RegistryController)
	return CopyViaJSON(dst, src)
}

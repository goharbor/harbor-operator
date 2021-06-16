package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (p *Portal) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Portal)

	return CopyViaJSON(dst, p)
}

func (p *Portal) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Portal)

	return CopyViaJSON(p, src)
}

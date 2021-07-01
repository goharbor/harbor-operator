package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Trivy{}

func (t *Trivy) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(t).To(dstRaw)
}

func (t *Trivy) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(t).From(srcRaw)
}

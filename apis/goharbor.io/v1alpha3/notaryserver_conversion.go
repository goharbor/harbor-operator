package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &NotaryServer{}

func (n *NotaryServer) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(n).To(dstRaw)
}

func (n *NotaryServer) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(n).From(srcRaw)
}

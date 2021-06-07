package v1alpha3

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &JobService{}

func (js *JobService) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(js).To(dstRaw)
}

func (js *JobService) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(js).From(srcRaw)
}

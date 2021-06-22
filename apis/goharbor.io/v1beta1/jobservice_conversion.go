package v1beta1

import (
	"github.com/goharbor/harbor-operator/pkg/convert"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (j *JobService) ConvertTo(dstRaw conversion.Hub) error {
	return convert.ConverterObject(j).To(dstRaw)
}

func (j *JobService) ConvertFrom(srcRaw conversion.Hub) error {
	return convert.ConverterObject(j).From(srcRaw)
}

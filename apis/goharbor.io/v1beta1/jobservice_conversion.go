package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (j *JobService) ConvertTo(dstRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(j, dstRaw)
}

func (j *JobService) ConvertFrom(srcRaw conversion.Hub) error {
	return ConvertViaUnstructuredCopy(srcRaw, j)
}

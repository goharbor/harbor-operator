package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (t *Trivy) ConvertTo(dstRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(t)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, dstRaw); err != nil {
		return err
	}

	dstRaw.(*v1alpha3.Trivy).APIVersion = GroupVersion.String()

	return nil
}

func (t *Trivy) ConvertFrom(srcRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(srcRaw)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, t); err != nil {
		return err
	}

	t.APIVersion = GroupVersion.String()

	return nil
}

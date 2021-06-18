package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (r *RegistryController) ConvertTo(dstRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, dstRaw); err != nil {
		return err
	}

	dstRaw.(*v1alpha3.RegistryController).APIVersion = v1alpha3.GroupVersion.String()

	return nil
}

func (r *RegistryController) ConvertFrom(srcRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(srcRaw)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, r); err != nil {
		return err
	}

	r.APIVersion = GroupVersion.String()

	return nil
}

package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (p *Portal) ConvertTo(dstRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(p)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, dstRaw); err != nil {
		return err
	}

	dstRaw.(*v1alpha3.Portal).APIVersion = v1alpha3.GroupVersion.String()

	return nil
}

func (p *Portal) ConvertFrom(srcRaw conversion.Hub) error {
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(srcRaw)
	if err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(src, p); err != nil {
		return err
	}

	p.APIVersion = GroupVersion.String()

	return nil
}

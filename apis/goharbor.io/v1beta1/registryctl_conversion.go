package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RegistryController) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.RegistryController)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_RegistryControllerSpec_To_v1alpha3_RegistryControllerSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *RegistryController) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.RegistryController)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_RegistryControllerSpec_To_v1beta1_RegistryControllerSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_RegistryControllerSpec_To_v1alpha3_RegistryControllerSpec(src *RegistryControllerSpec, dst *v1alpha3.RegistryControllerSpec) {

	dst.ComponentSpec = src.ComponentSpec
	dst.RegistryRef = src.RegistryRef
	dst.TLS = src.TLS

	dst.Authentication = v1alpha3.RegistryControllerAuthenticationSpec{
		CoreSecretRef:       src.Authentication.CoreSecretRef,
		JobServiceSecretRef: src.Authentication.JobServiceSecretRef,
	}

	dst.Log = v1alpha3.RegistryControllerLogSpec{
		Level: src.Log.Level,
	}
}

func Convert_v1alpha3_RegistryControllerSpec_To_v1beta1_RegistryControllerSpec(src *v1alpha3.RegistryControllerSpec, dst *RegistryControllerSpec) {

	dst.ComponentSpec = src.ComponentSpec
	dst.RegistryRef = src.RegistryRef
	dst.TLS = src.TLS

	dst.Authentication = RegistryControllerAuthenticationSpec{
		CoreSecretRef:       src.Authentication.CoreSecretRef,
		JobServiceSecretRef: src.Authentication.JobServiceSecretRef,
	}

	dst.Log = RegistryControllerLogSpec{
		Level: src.Log.Level,
	}
}

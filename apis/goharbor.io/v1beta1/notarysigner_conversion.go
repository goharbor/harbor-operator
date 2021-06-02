package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *NotarySigner) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.NotarySigner)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_NotarySignerSpec_To_v1alpha3_NotarySignerSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *NotarySigner) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.NotarySigner)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_NotarySignerSpec_To_v1beta1_NotarySignerSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_NotarySignerSpec_To_v1alpha3_NotarySignerSpec(src *NotarySignerSpec, dst *v1alpha3.NotarySignerSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.NotarySignerSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.MigrationEnabled = src.MigrationEnabled

	Convert_v1beta1_NotarySignerAuthenticationSpec_To_v1alpha3_NotarySignerAuthenticationSpec(&src.Authentication, &dst.Authentication)

	Convert_v1beta1_NotaryLoggingSpec_To_v1alpha3_NotaryLoggingSpec(&src.Logging, &dst.Logging)

	Convert_v1beta1_NotarySignerStorageSpec_To_v1alpha3_NotarySignerStorageSpec(&src.Storage, &dst.Storage)

}

func Convert_v1beta1_NotarySignerAuthenticationSpec_To_v1alpha3_NotarySignerAuthenticationSpec(src *NotarySignerAuthenticationSpec, dst *v1alpha3.NotarySignerAuthenticationSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.NotarySignerAuthenticationSpec{}
	}

	dst.CertificateRef = src.CertificateRef
}

func Convert_v1beta1_NotarySignerStorageSpec_To_v1alpha3_NotarySignerStorageSpec(src *NotarySignerStorageSpec, dst *v1alpha3.NotarySignerStorageSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.NotarySignerStorageSpec{}
	}

	dst.AliasesRef = src.AliasesRef

	Convert_v1beta1_NotaryStorageSpec_To_v1alpha3_NotaryStorageSpec(&src.NotaryStorageSpec, &dst.NotaryStorageSpec)

}

func Convert_v1alpha3_NotarySignerSpec_To_v1beta1_NotarySignerSpec(src *v1alpha3.NotarySignerSpec, dst *NotarySignerSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &NotarySignerSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.MigrationEnabled = src.MigrationEnabled

	Convert_v1alpha3_NotarySignerAuthenticationSpec_To_v1beta1_NotarySignerAuthenticationSpec(&src.Authentication, &dst.Authentication)

	Convert_v1alpha3_NotaryLoggingSpec_To_v1beta1_NotaryLoggingSpec(&src.Logging, &dst.Logging)

	Convert_v1alpha3_NotarySignerStorageSpec_To_v1beta1_NotarySignerStorageSpec(&src.Storage, &dst.Storage)

}

func Convert_v1alpha3_NotarySignerAuthenticationSpec_To_v1beta1_NotarySignerAuthenticationSpec(src *v1alpha3.NotarySignerAuthenticationSpec, dst *NotarySignerAuthenticationSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &NotarySignerAuthenticationSpec{}
	}

	dst.CertificateRef = src.CertificateRef
}

func Convert_v1alpha3_NotarySignerStorageSpec_To_v1beta1_NotarySignerStorageSpec(src *v1alpha3.NotarySignerStorageSpec, dst *NotarySignerStorageSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &NotarySignerStorageSpec{}
	}

	dst.AliasesRef = src.AliasesRef

	Convert_v1alpha3_NotaryStorageSpec_To_v1beta1_NotaryStorageSpec(&src.NotaryStorageSpec, &dst.NotaryStorageSpec)

}

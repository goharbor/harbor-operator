package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *NotaryServer) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.NotaryServer)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_NotaryServerSpec_To_v1alpha3_NotaryServerSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *NotaryServer) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.NotaryServer)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_NotaryServerSpec_To_v1beta1_NotaryServerSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_NotaryServerSpec_To_v1alpha3_NotaryServerSpec(src *NotaryServerSpec, dst *v1alpha3.NotaryServerSpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.TLS = src.TLS
	dst.MigrationEnabled = src.MigrationEnabled

	Convert_v1beta1_NotaryServerTrustServiceSpec_To_v1alpha3_NotaryServerTrustServiceSpec(&src.TrustService, &dst.TrustService)

	Convert_v1beta1_NotaryLoggingSpec_To_v1alpha3_NotaryLoggingSpec(&src.Logging, &dst.Logging)

	Convert_v1beta1_NotaryStorageSpec_To_v1alpha3_NotaryStorageSpec(&src.Storage, &dst.Storage)

	if src.Authentication != nil {
		dst.Authentication = &v1alpha3.NotaryServerAuthSpec{}
		Convert_v1beta1_NotaryServerAuthSpec_To_v1alpha3_NotaryServerAuthSpec(src.Authentication, dst.Authentication)
	}

}

func Convert_v1beta1_NotaryServerTrustServiceSpec_To_v1alpha3_NotaryServerTrustServiceSpec(src *NotaryServerTrustServiceSpec, dst *v1alpha3.NotaryServerTrustServiceSpec) {
	dst.Remote = &v1alpha3.NotaryServerTrustServiceRemoteSpec{
		Host:           src.Remote.Host,
		Port:           src.Remote.Port,
		KeyAlgorithm:   src.Remote.KeyAlgorithm,
		CertificateRef: src.Remote.CertificateRef,
	}
}

func Convert_v1beta1_NotaryServerAuthSpec_To_v1alpha3_NotaryServerAuthSpec(src *NotaryServerAuthSpec, dst *v1alpha3.NotaryServerAuthSpec) {
	dst.Token = v1alpha3.NotaryServerAuthTokenSpec{
		Realm:          src.Token.Realm,
		Service:        src.Token.Service,
		Issuer:         src.Token.Issuer,
		CertificateRef: src.Token.CertificateRef,
		AutoRedirect:   src.Token.AutoRedirect,
	}
}

func Convert_v1alpha3_NotaryServerSpec_To_v1beta1_NotaryServerSpec(src *v1alpha3.NotaryServerSpec, dst *NotaryServerSpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.TLS = src.TLS
	dst.MigrationEnabled = src.MigrationEnabled

	Convert_v1alpha3_NotaryServerTrustServiceSpec_To_v1beta1_NotaryServerTrustServiceSpec(&src.TrustService, &dst.TrustService)

	Convert_v1alpha3_NotaryLoggingSpec_To_v1beta1_NotaryLoggingSpec(&src.Logging, &dst.Logging)

	Convert_v1alpha3_NotaryStorageSpec_To_v1beta1_NotaryStorageSpec(&src.Storage, &dst.Storage)

	if src.Authentication != nil {
		dst.Authentication = &NotaryServerAuthSpec{}
		Convert_v1alpha3_NotaryServerAuthSpec_To_v1beta1_NotaryServerAuthSpec(src.Authentication, dst.Authentication)
	}
}

func Convert_v1alpha3_NotaryServerTrustServiceSpec_To_v1beta1_NotaryServerTrustServiceSpec(src *v1alpha3.NotaryServerTrustServiceSpec, dst *NotaryServerTrustServiceSpec) {
	dst.Remote = &NotaryServerTrustServiceRemoteSpec{
		Host:           src.Remote.Host,
		Port:           src.Remote.Port,
		KeyAlgorithm:   src.Remote.KeyAlgorithm,
		CertificateRef: src.Remote.CertificateRef,
	}
}

func Convert_v1alpha3_NotaryServerAuthSpec_To_v1beta1_NotaryServerAuthSpec(src *v1alpha3.NotaryServerAuthSpec, dst *NotaryServerAuthSpec) {
	dst.Token = NotaryServerAuthTokenSpec{
		Realm:          src.Token.Realm,
		Service:        src.Token.Service,
		Issuer:         src.Token.Issuer,
		CertificateRef: src.Token.CertificateRef,
		AutoRedirect:   src.Token.AutoRedirect,
	}
}

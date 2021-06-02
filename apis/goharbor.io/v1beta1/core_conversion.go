package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Core) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Core)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_CoreSpec_To_v1alpha3_CoreSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *Core) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Core)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_CoreSpec_To_v1beta1_CoreSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_CoreSpec_To_v1alpha3_CoreSpec(src *CoreSpec, dst *v1alpha3.CoreSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{
		CertificateRefs: src.CertificateRefs,
	}
	dst.Proxy = src.Proxy
	dst.ExternalEndpoint = src.ExternalEndpoint
	dst.Metrics = src.Metrics
	dst.ConfigExpiration = src.ConfigExpiration
	dst.CSRFKeyRef = src.CSRFKeyRef

	Convert_v1beta1_CoreConfig_To_v1alpha3_CoreConfig(&src.CoreConfig, &dst.CoreConfig)

	Convert_v1beta1_CoreHTTPSpec_To_v1alpha3_CoreHTTPSpec(&src.HTTP, &dst.HTTP)

	Convert_v1beta1_CoreComponentsSpec_To_v1alpha3_CoreComponentsSpec(&src.Components, &dst.Components)

	Convert_v1beta1_CoreLogSpec_To_v1alpha3_CoreLogSpec(&src.Log, &dst.Log)

	Convert_v1beta1_CoreDatabaseSpec_To_v1alpha3_CoreDatabaseSpec(&src.Database, &dst.Database)

	Convert_v1beta1_CoreRedisSpec_To_v1alpha3_CoreRedisSpec(&src.Redis, &dst.Redis)

}

func Convert_v1beta1_CoreConfig_To_v1alpha3_CoreConfig(src *CoreConfig, dst *v1alpha3.CoreConfig) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreConfig{}
	}

	dst.SecretRef = src.SecretRef
	dst.AdminInitialPasswordRef = src.AdminInitialPasswordRef
	dst.AuthenticationMode = src.AuthenticationMode
	dst.PublicCertificateRef = src.PublicCertificateRef
}

func Convert_v1beta1_CoreHTTPSpec_To_v1alpha3_CoreHTTPSpec(src *CoreHTTPSpec, dst *v1alpha3.CoreHTTPSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreHTTPSpec{}
	}

	dst.GZip = src.GZip
}

func Convert_v1beta1_CoreComponentsSpec_To_v1alpha3_CoreComponentsSpec(src *CoreComponentsSpec, dst *v1alpha3.CoreComponentsSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsSpec{}
	}

	dst.TLS = src.TLS

	Convert_v1beta1_CoreComponentsJobServiceSpec_To_v1alpha3_CoreComponentsJobServiceSpec(&src.JobService, &dst.JobService)

	Convert_v1beta1_CoreComponentsPortalSpec_To_v1alpha3_CoreComponentsPortalSpec(&src.Portal, &dst.Portal)

	Convert_v1beta1_CoreComponentsRegistrySpec_To_v1alpha3_CoreComponentsRegistrySpec(&src.Registry, &dst.Registry)

	Convert_v1beta1_CoreComponentsTokenServiceSpec_To_v1alpha3_CoreComponentsTokenServiceSpec(&src.TokenService, &dst.TokenService)

	Convert_v1beta1_CoreComponentsTrivySpec_To_v1alpha3_CoreComponentsTrivySpec(src.Trivy, dst.Trivy)

	Convert_v1beta1_CoreComponentsChartRepositorySpec_To_v1alpha3_CoreComponentsChartRepositorySpec(src.ChartRepository, dst.ChartRepository)

	Convert_v1beta1_CoreComponentsNotaryServerSpec_To_v1alpha3_CoreComponentsNotaryServerSpec(src.NotaryServer, dst.NotaryServer)

}

func Convert_v1beta1_CoreComponentsJobServiceSpec_To_v1alpha3_CoreComponentsJobServiceSpec(src *CoreComponentsJobServiceSpec, dst *v1alpha3.CoreComponentsJobServiceSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsJobServiceSpec{}
	}

	dst.SecretRef = src.SecretRef
	dst.URL = src.URL
}

func Convert_v1beta1_CoreComponentsPortalSpec_To_v1alpha3_CoreComponentsPortalSpec(src *CoreComponentPortalSpec, dst *v1alpha3.CoreComponentPortalSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentPortalSpec{}
	}

	dst.URL = src.URL
}

func Convert_v1beta1_CoreComponentsRegistrySpec_To_v1alpha3_CoreComponentsRegistrySpec(src *CoreComponentsRegistrySpec, dst *v1alpha3.CoreComponentsRegistrySpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsRegistrySpec{}
	}

	dst.Redis = src.Redis
	dst.Sync = src.Sync
	dst.StorageProviderName = src.StorageProviderName

	dst.RegistryControllerConnectionSpec = v1alpha3.RegistryControllerConnectionSpec{
		RegistryURL:   src.RegistryURL,
		ControllerURL: src.RegistryURL,
		Credentials: v1alpha3.CoreComponentsRegistryCredentialsSpec{
			Username:    src.Credentials.Username,
			PasswordRef: src.Credentials.PasswordRef,
		},
	}
}

func Convert_v1beta1_CoreComponentsTokenServiceSpec_To_v1alpha3_CoreComponentsTokenServiceSpec(src *CoreComponentsTokenServiceSpec, dst *v1alpha3.CoreComponentsTokenServiceSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsTokenServiceSpec{}
	}

	dst.URL = src.URL
	dst.CertificateRef = src.URL
}

func Convert_v1beta1_CoreComponentsTrivySpec_To_v1alpha3_CoreComponentsTrivySpec(src *CoreComponentsTrivySpec, dst *v1alpha3.CoreComponentsTrivySpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsTrivySpec{}
	}

	dst.URL = src.URL
	dst.AdapterURL = src.AdapterURL
}

func Convert_v1beta1_CoreComponentsChartRepositorySpec_To_v1alpha3_CoreComponentsChartRepositorySpec(src *CoreComponentsChartRepositorySpec, dst *v1alpha3.CoreComponentsChartRepositorySpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsChartRepositorySpec{}
	}

	dst.URL = src.URL
	dst.AbsoluteURL = src.AbsoluteURL
	dst.CacheDriver = src.CacheDriver
}

func Convert_v1beta1_CoreComponentsNotaryServerSpec_To_v1alpha3_CoreComponentsNotaryServerSpec(src *CoreComponentsNotaryServerSpec, dst *v1alpha3.CoreComponentsNotaryServerSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreComponentsNotaryServerSpec{}
	}

	dst.URL = src.URL
}

func Convert_v1beta1_CoreLogSpec_To_v1alpha3_CoreLogSpec(src *CoreLogSpec, dst *v1alpha3.CoreLogSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreLogSpec{}
	}

	dst.Level = src.Level
}

func Convert_v1beta1_CoreDatabaseSpec_To_v1alpha3_CoreDatabaseSpec(src *CoreDatabaseSpec, dst *v1alpha3.CoreDatabaseSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreDatabaseSpec{}
	}

	dst.EncryptionKeyRef = src.EncryptionKeyRef
	dst.MaxIdleConnections = src.MaxIdleConnections
	dst.MaxOpenConnections = src.MaxOpenConnections
	dst.PostgresConnectionWithParameters = src.PostgresConnectionWithParameters
}

func Convert_v1beta1_CoreRedisSpec_To_v1alpha3_CoreRedisSpec(src *CoreRedisSpec, dst *v1alpha3.CoreRedisSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.CoreRedisSpec{}
	}

	dst.IdleTimeout = src.IdleTimeout
	dst.RedisConnection = src.RedisConnection

}

func Convert_v1alpha3_CoreSpec_To_v1beta1_CoreSpec(src *v1alpha3.CoreSpec, dst *CoreSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateRefs,
	}
	dst.Proxy = src.Proxy
	dst.ExternalEndpoint = src.ExternalEndpoint
	dst.Metrics = src.Metrics
	dst.ConfigExpiration = src.ConfigExpiration
	dst.CSRFKeyRef = src.CSRFKeyRef

	Convert_v1alpha3_CoreConfig_To_v1beta1_CoreConfig(&src.CoreConfig, &dst.CoreConfig)

	Convert_v1alpha3_CoreHTTPSpec_To_v1beta1_CoreHTTPSpec(&src.HTTP, &dst.HTTP)

	Convert_v1alpha3_CoreComponentsSpec_To_v1beta1_CoreComponentsSpec(&src.Components, &dst.Components)

	Convert_v1alpha3_CoreLogSpec_To_v1beta1_CoreLogSpec(&src.Log, &dst.Log)

	Convert_v1alpha3_CoreDatabaseSpec_To_v1beta1_CoreDatabaseSpec(&src.Database, &dst.Database)

	Convert_v1alpha3_CoreRedisSpec_To_v1beta1_CoreRedisSpec(&src.Redis, &dst.Redis)

}

func Convert_v1alpha3_CoreConfig_To_v1beta1_CoreConfig(src *v1alpha3.CoreConfig, dst *CoreConfig) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreConfig{}
	}

	dst.SecretRef = src.SecretRef
	dst.AdminInitialPasswordRef = src.AdminInitialPasswordRef
	dst.AuthenticationMode = src.AuthenticationMode
	dst.PublicCertificateRef = src.PublicCertificateRef
}

func Convert_v1alpha3_CoreHTTPSpec_To_v1beta1_CoreHTTPSpec(src *v1alpha3.CoreHTTPSpec, dst *CoreHTTPSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreHTTPSpec{}
	}

	dst.GZip = src.GZip
}

func Convert_v1alpha3_CoreComponentsSpec_To_v1beta1_CoreComponentsSpec(src *v1alpha3.CoreComponentsSpec, dst *CoreComponentsSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsSpec{}
	}

	dst.TLS = src.TLS

	Convert_v1alpha3_CoreComponentsJobServiceSpec_To_v1beta1_CoreComponentsJobServiceSpec(&src.JobService, &dst.JobService)

	Convert_v1alpha3_CoreComponentsPortalSpec_To_v1beta1_CoreComponentsPortalSpec(&src.Portal, &dst.Portal)

	Convert_v1alpha3_CoreComponentsRegistrySpec_To_v1beta1_CoreComponentsRegistrySpec(&src.Registry, &dst.Registry)

	Convert_v1alpha3_CoreComponentsTokenServiceSpec_To_v1beta1_CoreComponentsTokenServiceSpec(&src.TokenService, &dst.TokenService)

	Convert_v1alpha3_CoreComponentsTrivySpec_To_v1beta1_CoreComponentsTrivySpec(src.Trivy, dst.Trivy)

	Convert_v1alpha3_CoreComponentsChartRepositorySpec_To_v1beta1_CoreComponentsChartRepositorySpec(src.ChartRepository, dst.ChartRepository)

	Convert_v1alpha3_CoreComponentsNotaryServerSpec_To_v1beta1_CoreComponentsNotaryServerSpec(src.NotaryServer, dst.NotaryServer)

}

func Convert_v1alpha3_CoreComponentsJobServiceSpec_To_v1beta1_CoreComponentsJobServiceSpec(src *v1alpha3.CoreComponentsJobServiceSpec, dst *CoreComponentsJobServiceSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsJobServiceSpec{}
	}

	dst.SecretRef = src.SecretRef
	dst.URL = src.URL
}

func Convert_v1alpha3_CoreComponentsPortalSpec_To_v1beta1_CoreComponentsPortalSpec(src *v1alpha3.CoreComponentPortalSpec, dst *CoreComponentPortalSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentPortalSpec{}
	}

	dst.URL = src.URL
}

func Convert_v1alpha3_CoreComponentsRegistrySpec_To_v1beta1_CoreComponentsRegistrySpec(src *v1alpha3.CoreComponentsRegistrySpec, dst *CoreComponentsRegistrySpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsRegistrySpec{}
	}

	dst.Redis = src.Redis
	dst.Sync = src.Sync
	dst.StorageProviderName = src.StorageProviderName

	dst.RegistryControllerConnectionSpec = RegistryControllerConnectionSpec{
		RegistryURL:   src.RegistryURL,
		ControllerURL: src.RegistryURL,
		Credentials: CoreComponentsRegistryCredentialsSpec{
			Username:    src.Credentials.Username,
			PasswordRef: src.Credentials.PasswordRef,
		},
	}
}

func Convert_v1alpha3_CoreComponentsTokenServiceSpec_To_v1beta1_CoreComponentsTokenServiceSpec(src *v1alpha3.CoreComponentsTokenServiceSpec, dst *CoreComponentsTokenServiceSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsTokenServiceSpec{}
	}

	dst.URL = src.URL
	dst.CertificateRef = src.URL
}

func Convert_v1alpha3_CoreComponentsTrivySpec_To_v1beta1_CoreComponentsTrivySpec(src *v1alpha3.CoreComponentsTrivySpec, dst *CoreComponentsTrivySpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsTrivySpec{}
	}

	dst.URL = src.URL
	dst.AdapterURL = src.AdapterURL
}

func Convert_v1alpha3_CoreComponentsChartRepositorySpec_To_v1beta1_CoreComponentsChartRepositorySpec(src *v1alpha3.CoreComponentsChartRepositorySpec, dst *CoreComponentsChartRepositorySpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsChartRepositorySpec{}
	}

	dst.URL = src.URL
	dst.AbsoluteURL = src.AbsoluteURL
	dst.CacheDriver = src.CacheDriver
}

func Convert_v1alpha3_CoreComponentsNotaryServerSpec_To_v1beta1_CoreComponentsNotaryServerSpec(src *v1alpha3.CoreComponentsNotaryServerSpec, dst *CoreComponentsNotaryServerSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreComponentsNotaryServerSpec{}
	}

	dst.URL = src.URL
}

func Convert_v1alpha3_CoreLogSpec_To_v1beta1_CoreLogSpec(src *v1alpha3.CoreLogSpec, dst *CoreLogSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreLogSpec{}
	}

	dst.Level = src.Level
}

func Convert_v1alpha3_CoreDatabaseSpec_To_v1beta1_CoreDatabaseSpec(src *v1alpha3.CoreDatabaseSpec, dst *CoreDatabaseSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreDatabaseSpec{}
	}

	dst.EncryptionKeyRef = src.EncryptionKeyRef
	dst.MaxIdleConnections = src.MaxIdleConnections
	dst.MaxOpenConnections = src.MaxOpenConnections
	dst.PostgresConnectionWithParameters = src.PostgresConnectionWithParameters
}

func Convert_v1alpha3_CoreRedisSpec_To_v1beta1_CoreRedisSpec(src *v1alpha3.CoreRedisSpec, dst *CoreRedisSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &CoreRedisSpec{}
	}

	dst.IdleTimeout = src.IdleTimeout
	dst.RedisConnection = src.RedisConnection

}

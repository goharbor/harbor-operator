package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Trivy) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Trivy)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_TrivySpec_To_v1alpha3_TrivySpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *Trivy) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Trivy)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_TrivySpec_To_v1beta1_TrivySpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_TrivySpec_To_v1alpha3_TrivySpec(src *TrivySpec, dst *v1alpha3.TrivySpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.TrivyVulnerabilityTypes = src.TrivyVulnerabilityTypes
	dst.TrivySeverityTypes = src.TrivySeverityTypes

	dst.CertificateInjection = v1alpha3.CertificateInjection{
		CertificateRefs: src.CertificateRefs,
	}

	dst.Proxy = src.Proxy

	dst.Log = v1alpha3.TrivyLogSpec{
		Level: src.Log.Level,
	}

	dst.Update = v1alpha3.TrivyUpdateSpec{
		Skip:           src.Update.Skip,
		GithubTokenRef: src.Update.GithubTokenRef,
	}

	Convert_v1beta1_TrivyServerSpec_To_v1alpha3_TrivyServerSpec(&src.Server, &dst.Server)

	Convert_v1beta1_TrivyStorageSpec_To_v1alpha3_TrivyStorageSpec(&src.Storage, &dst.Storage)

	Convert_v1beta1_TrivyRedisSpec_To_v1alpha3_TrivyRedisSpec(&src.Redis, &dst.Redis)
}

func Convert_v1beta1_TrivyServerSpec_To_v1alpha3_TrivyServerSpec(src *TrivyServerSpec, dst *v1alpha3.TrivyServerSpec) {
	dst.TLS = src.TLS
	dst.ClientCertificateAuthorityRefs = src.ClientCertificateAuthorityRefs
	dst.TokenServiceCertificateAuthorityRefs = src.TokenServiceCertificateAuthorityRefs
	dst.IdleTimeout = src.IdleTimeout
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
	dst.IgnoreUnfixed = src.IgnoreUnfixed
	dst.DebugMode = src.DebugMode
	dst.Insecure = src.Insecure

	if src.Proxy != nil {
		dst.Proxy = &v1alpha3.TrivyServerProxySpec{
			URL:     src.Proxy.URL,
			NoProxy: src.Proxy.NoProxy,
		}
	}
}

func Convert_v1beta1_TrivyRedisSpec_To_v1alpha3_TrivyRedisSpec(src *TrivyRedisSpec, dst *v1alpha3.TrivyRedisSpec) {
	dst.RedisConnection = src.RedisConnection
	dst.Namespace = src.Namespace

	dst.Jobs = v1alpha3.TrivyRedisJobsSpec{
		ScanTTL:   src.Jobs.ScanTTL,
		Namespace: src.Jobs.Namespace,
	}

	dst.Pool = v1alpha3.TrivyRedisPoolSpec{
		MaxActive:         src.Pool.MaxActive,
		MaxIdle:           src.Pool.MaxIdle,
		IdleTimeout:       src.Pool.IdleTimeout,
		ConnectionTimeout: src.Pool.ConnectionTimeout,
		ReadTimeout:       src.Pool.ReadTimeout,
		WriteTimeout:      src.Pool.WriteTimeout,
	}
}

func Convert_v1beta1_TrivyStorageSpec_To_v1alpha3_TrivyStorageSpec(src *TrivyStorageSpec, dst *v1alpha3.TrivyStorageSpec) {
	dst.Reports = v1alpha3.TrivyStorageVolumeSpec{
		VolumeSource: src.Reports.VolumeSource,
		Prefix:       src.Reports.Prefix,
	}

	dst.Cache = v1alpha3.TrivyStorageVolumeSpec{
		VolumeSource: src.Cache.VolumeSource,
		Prefix:       src.Cache.Prefix,
	}
}

func Convert_v1alpha3_TrivySpec_To_v1beta1_TrivySpec(src *v1alpha3.TrivySpec, dst *TrivySpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.TrivyVulnerabilityTypes = src.TrivyVulnerabilityTypes
	dst.TrivySeverityTypes = src.TrivySeverityTypes

	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateRefs,
	}

	dst.Proxy = src.Proxy

	dst.Log = TrivyLogSpec{
		Level: src.Log.Level,
	}

	dst.Update = TrivyUpdateSpec{
		Skip:           src.Update.Skip,
		GithubTokenRef: src.Update.GithubTokenRef,
	}

	Convert_v1alpha3_TrivyServerSpec_To_v1beta1_TrivyServerSpec(&src.Server, &dst.Server)

	Convert_v1alpha3_TrivyStorageSpec_To_v1beta1_TrivyStorageSpec(&src.Storage, &dst.Storage)

	Convert_v1alpha3_TrivyRedisSpec_To_v1beta1_TrivyRedisSpec(&src.Redis, &dst.Redis)
}

func Convert_v1alpha3_TrivyServerSpec_To_v1beta1_TrivyServerSpec(src *v1alpha3.TrivyServerSpec, dst *TrivyServerSpec) {
	dst.TLS = src.TLS
	dst.ClientCertificateAuthorityRefs = src.ClientCertificateAuthorityRefs
	dst.TokenServiceCertificateAuthorityRefs = src.TokenServiceCertificateAuthorityRefs
	dst.IdleTimeout = src.IdleTimeout
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
	dst.IgnoreUnfixed = src.IgnoreUnfixed
	dst.DebugMode = src.DebugMode
	dst.Insecure = src.Insecure

	if src.Proxy != nil {
		dst.Proxy = &TrivyServerProxySpec{
			URL:     src.Proxy.URL,
			NoProxy: src.Proxy.NoProxy,
		}
	}
}

func Convert_v1alpha3_TrivyRedisSpec_To_v1beta1_TrivyRedisSpec(src *v1alpha3.TrivyRedisSpec, dst *TrivyRedisSpec) {
	dst.RedisConnection = src.RedisConnection
	dst.Namespace = src.Namespace

	dst.Jobs = TrivyRedisJobsSpec{
		ScanTTL:   src.Jobs.ScanTTL,
		Namespace: src.Jobs.Namespace,
	}

	dst.Pool = TrivyRedisPoolSpec{
		MaxActive:         src.Pool.MaxActive,
		MaxIdle:           src.Pool.MaxIdle,
		IdleTimeout:       src.Pool.IdleTimeout,
		ConnectionTimeout: src.Pool.ConnectionTimeout,
		ReadTimeout:       src.Pool.ReadTimeout,
		WriteTimeout:      src.Pool.WriteTimeout,
	}
}

func Convert_v1alpha3_TrivyStorageSpec_To_v1beta1_TrivyStorageSpec(src *v1alpha3.TrivyStorageSpec, dst *TrivyStorageSpec) {
	dst.Reports = TrivyStorageVolumeSpec{
		VolumeSource: src.Reports.VolumeSource,
		Prefix:       src.Reports.Prefix,
	}

	dst.Cache = TrivyStorageVolumeSpec{
		VolumeSource: src.Cache.VolumeSource,
		Prefix:       src.Cache.Prefix,
	}
}
